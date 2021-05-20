// Copyright 2012 Google, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

// The pcapdump binary implements a tcpdump-like command line tool with gopacket
// using pcap as a backend data collection mechanism.
package tap

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/examples/util"
	"github.com/google/gopacket/ip4defrag"
	"github.com/google/gopacket/layers" // pulls in all layers decoders
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/reassembly"
)

const AppPortsEnvVar = "APP_PORTS"
const OutPortEnvVar = "WEB_SOCKET_PORT"
const maxHTTP2DataLenEnvVar = "HTTP2_DATA_SIZE_LIMIT"
const hostModeEnvVar = "HOST_MODE"
// default is 1MB, more than the max size accepted by collector and traffic-dumper
const maxHTTP2DataLenDefault = 1 * 1024 * 1024
const cleanPeriod = time.Second * 10
const outboundThrottleCacheExpiryPeriod = time.Minute * 15
var remoteOnlyOutboundPorts = []int { 80, 443 }

func parseAppPorts(appPortsList string) []int {
	ports := make([]int, 0)
	for _, portStr := range strings.Split(appPortsList, ",") {
		parsedInt, parseError := strconv.Atoi(portStr)
		if parseError != nil {
			fmt.Println("Provided app port ", portStr, " is not a valid number!")
		} else {
			ports = append(ports, parsedInt)
		}
	}
	return ports
}

func parseHostAppAddresses(hostAppAddressesString string) []string {
	if len(hostAppAddressesString) == 0 {
		return []string{}
	}
	return strings.Split(hostAppAddressesString, ",")
}

var maxcount = flag.Int("c", -1, "Only grab this many packets, then exit")
var decoder = flag.String("decoder", "", "Name of the decoder to use (default: guess from capture)")
var statsevery = flag.Int("stats", 60, "Output statistics every N seconds")
var lazy = flag.Bool("lazy", false, "If true, do lazy decoding")
var nodefrag = flag.Bool("nodefrag", false, "If true, do not do IPv4 defrag")
var checksum = flag.Bool("checksum", false, "Check TCP checksum")  // global
var nooptcheck = flag.Bool("nooptcheck", true, "Do not check TCP options (useful to ignore MSS on captures with TSO)")  // global
var ignorefsmerr = flag.Bool("ignorefsmerr", true, "Ignore TCP FSM errors")  // global
var allowmissinginit = flag.Bool("allowmissinginit", true, "Support streams without SYN/SYN+ACK/ACK sequence")  // global
var verbose = flag.Bool("verbose", false, "Be verbose")
var debug = flag.Bool("debug", false, "Display debug information")
var quiet = flag.Bool("quiet", false, "Be quiet regarding errors")

// http
var nohttp = flag.Bool("nohttp", false, "Disable HTTP parsing")
var output = flag.String("output", "", "Path to create file for HTTP 200 OK responses")
var writeincomplete = flag.Bool("writeincomplete", false, "Write incomplete response")

var hexdump = flag.Bool("dump", false, "Dump HTTP request/response as hex")  // global
var hexdumppkt = flag.Bool("dumppkt", false, "Dump packet as hex")

// capture
var iface = flag.String("i", "en0", "Interface to read packets from")
var fname = flag.String("r", "", "Filename to read from, overrides -i")
var snaplen = flag.Int("s", 65536, "Snap length (number of bytes max to read per packet")
var tstype = flag.String("timestamp_type", "", "Type of timestamps to use")
var promisc = flag.Bool("promisc", true, "Set promiscuous mode")
var anydirection = flag.Bool("anydirection", false, "Capture http requests to other hosts")
var staleTimeoutSeconds = flag.Int("staletimout", 120, "Max time in seconds to keep connections which don't transmit data")
var hostAppAddressesString = flag.String("targets", "", "Comma separated list of ip:ports to tap")

var memprofile = flag.String("memprofile", "", "Write memory profile")

// output
var dumpToHar = flag.Bool("hardump", false, "Dump traffic to har files")
var HarOutputDir = flag.String("hardir", "", "Directory in which to store output har files")
var harEntriesPerFile = flag.Int("harentriesperfile", 200, "Number of max number of har entries to store in each file")

var reqResMatcher = createResponseRequestMatcher()  // global
var statsTracker = StatsTracker{}

// global
var stats struct {
	ipdefrag            int
	missedBytes         int
	pkt                 int
	sz                  int
	totalsz             int
	rejectFsm           int
	rejectOpt           int
	rejectConnFsm       int
	reassembled         int
	outOfOrderBytes     int
	outOfOrderPackets   int
	biggestChunkBytes   int
	biggestChunkPackets int
	overlapBytes        int
	overlapPackets      int
}

type CollectorMessage struct {
	MessageType string
	Ports *[]int `json:"ports,omitempty"`
	Addresses *[]string `json:"addresses,omitempty"`
}

var outputLevel int
var errorsMap map[string]uint
var errorsMapMutex sync.Mutex
var nErrors uint
var appPorts []int            // global
var ownIps []string           //global
var hostMode bool             //global
var HostAppAddresses []string //global

/* minOutputLevel: Error will be printed only if outputLevel is above this value
 * t:              key for errorsMap (counting errors)
 * s, a:           arguments fmt.Printf
 * Note:           Too bad for perf that a... is evaluated
 */
func logError(minOutputLevel int, t string, s string, a ...interface{}) {
	errorsMapMutex.Lock()
	nErrors++
	nb, _ := errorsMap[t]
	errorsMap[t] = nb + 1
	errorsMapMutex.Unlock()
	if outputLevel >= minOutputLevel {
		formatStr := fmt.Sprintf("%s: %s", t, s)
		fmt.Printf(formatStr, a...)
	}
}
func Error(t string, s string, a ...interface{}) {
	logError(0, t, s, a...)
}
func SilentError(t string, s string, a ...interface{}) {
	logError(2, t, s, a...)
}
func Info(s string, a ...interface{}) {
	if outputLevel >= 1 {
		fmt.Printf(s, a...)
	}
}
func Debug(s string, a ...interface{}) {
	if outputLevel >= 2 {
		fmt.Printf(s, a...)
	}
}

func inArrayInt(arr []int, valueToCheck int) bool {
	for _, value := range arr {
		if value == valueToCheck {
			return true
		}
	}
	return false
}

func inArrayString(arr []string, valueToCheck string) bool {
	for _, value := range arr {
		if value == valueToCheck {
			return true
		}
	}
	return false
}

/*
 * The assembler context
 */
type Context struct {
	CaptureInfo gopacket.CaptureInfo
}

func (c *Context) GetCaptureInfo() gopacket.CaptureInfo {
	return c.CaptureInfo
}

func StartPassiveTapper() <-chan *OutputChannelItem {
	var harWriter *HarWriter
	if *dumpToHar {
		harWriter = NewHarWriter(*HarOutputDir, *harEntriesPerFile)
	}

	go startPassiveTapper(harWriter)

	if harWriter != nil {
		return harWriter.OutChan
	}

	return nil
}

func startPassiveTapper(harWriter *HarWriter) {
	defer util.Run()()
	if *debug {
		outputLevel = 2
	} else if *verbose {
		outputLevel = 1
	} else if *quiet {
		outputLevel = -1
	}
	errorsMap = make(map[string]uint)

	if localhostIPs, err := getLocalhostIPs(); err != nil {
		// TODO: think this over
		fmt.Println("Failed to get self IP addresses")
		Error("Getting-Self-Address", "Error getting self ip address: %s (%v,%+v)\n", err, err, err)
		ownIps = make([]string, 0)
	} else {
		ownIps = localhostIPs
	}

	appPortsStr := os.Getenv(AppPortsEnvVar)
	if appPortsStr == "" {
		fmt.Println("Received empty/no APP_PORTS env var! only listening to http on port 80!")
		appPorts = make([]int, 0)
	} else {
		appPorts = parseAppPorts(appPortsStr)
	}
	//HostAppAddresses = parseHostAppAddresses(*hostAppAddressesString)
	fmt.Println("Filtering for the following addresses:", HostAppAddresses)
	tapOutputPort := os.Getenv(OutPortEnvVar)
	if tapOutputPort == "" {
		fmt.Println("Received empty/no WEB_SOCKET_PORT env var! falling back to port 8080")
		tapOutputPort = "8080"
	}
	envVal := os.Getenv(maxHTTP2DataLenEnvVar)
	if envVal == "" {
		fmt.Println("Received empty/no HTTP2_DATA_SIZE_LIMIT env var! falling back to", maxHTTP2DataLenDefault)
		maxHTTP2DataLen = maxHTTP2DataLenDefault
	} else {
		if convertedInt, err := strconv.Atoi(envVal); err != nil {
			fmt.Println("Received invalid HTTP2_DATA_SIZE_LIMIT env var! falling back to", maxHTTP2DataLenDefault)
			maxHTTP2DataLen = maxHTTP2DataLenDefault
		} else {
			fmt.Println("Received HTTP2_DATA_SIZE_LIMIT env var:", maxHTTP2DataLenDefault)
			maxHTTP2DataLen = convertedInt
		}
	}
	hostMode = os.Getenv(hostModeEnvVar) == "1"

	fmt.Printf("App Ports: %v\n", appPorts)
	fmt.Printf("Tap output websocket port: %s\n", tapOutputPort)

	var onCollectorMessage = func(message []byte) {
		var parsedMessage CollectorMessage
		err := json.Unmarshal(message, &parsedMessage)
		if err == nil {

			if parsedMessage.MessageType == "setPorts" {
				Debug("Got message from collector. Type: %s, Ports: %v\n", parsedMessage.MessageType, parsedMessage.Ports)
				appPorts = *parsedMessage.Ports
			} else if parsedMessage.MessageType == "setAddresses" {
				Debug("Got message from collector. Type: %s, IPs: %v\n", parsedMessage.MessageType, parsedMessage.Addresses)
				HostAppAddresses = *parsedMessage.Addresses
			}
		} else {
			Error("Collector-Message-Parsing", "Error parsing message from collector: %s (%v,%+v)\n", err, err, err)
		}
	}

	go startOutputServer(tapOutputPort, onCollectorMessage)

	var handle *pcap.Handle
	var err error
	if *fname != "" {
		if handle, err = pcap.OpenOffline(*fname); err != nil {
			log.Fatal("PCAP OpenOffline error:", err)
		}
	} else {
		// This is a little complicated because we want to allow all possible options
		// for creating the packet capture handle... instead of all this you can
		// just call pcap.OpenLive if you want a simple handle.
		inactive, err := pcap.NewInactiveHandle(*iface)
		if err != nil {
			log.Fatalf("could not create: %v", err)
		}
		defer inactive.CleanUp()
		if err = inactive.SetSnapLen(*snaplen); err != nil {
			log.Fatalf("could not set snap length: %v", err)
		} else if err = inactive.SetPromisc(*promisc); err != nil {
			log.Fatalf("could not set promisc mode: %v", err)
		} else if err = inactive.SetTimeout(time.Second); err != nil {
			log.Fatalf("could not set timeout: %v", err)
		}
		if *tstype != "" {
			if t, err := pcap.TimestampSourceFromString(*tstype); err != nil {
				log.Fatalf("Supported timestamp types: %v", inactive.SupportedTimestamps())
			} else if err := inactive.SetTimestampSource(t); err != nil {
				log.Fatalf("Supported timestamp types: %v", inactive.SupportedTimestamps())
			}
		}
		if handle, err = inactive.Activate(); err != nil {
			log.Fatal("PCAP Activate error:", err)
		}
		defer handle.Close()
	}
	if len(flag.Args()) > 0 {
		bpffilter := strings.Join(flag.Args(), " ")
		Info("Using BPF filter %q\n", bpffilter)
		if err = handle.SetBPFFilter(bpffilter); err != nil {
			log.Fatal("BPF filter error:", err)
		}
	}

	if *dumpToHar {
		harWriter.Start()
		defer harWriter.Stop()
	}

	var dec gopacket.Decoder
	var ok bool
	decoderName := *decoder
	if decoderName == "" {
		decoderName = fmt.Sprintf("%s", handle.LinkType())
	}
	if dec, ok = gopacket.DecodersByLayerName[decoderName]; !ok {
		log.Fatalln("No decoder named", decoderName)
	}
	source := gopacket.NewPacketSource(handle, dec)
	source.Lazy = *lazy
	source.NoCopy = true
	Info("Starting to read packets\n")
	count := 0
	bytes := int64(0)
	start := time.Now()
	defragger := ip4defrag.NewIPv4Defragmenter()

	streamFactory := &tcpStreamFactory{doHTTP: !*nohttp, harWriter: harWriter}
	streamPool := reassembly.NewStreamPool(streamFactory)
	assembler := reassembly.NewAssembler(streamPool)
	var assemblerMutex sync.Mutex

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	staleConnectionTimeout := time.Second * time.Duration(*staleTimeoutSeconds)
	cleaner := Cleaner{
		assembler: assembler,
		assemblerMutex: &assemblerMutex,
		matcher: &reqResMatcher,
		cleanPeriod: cleanPeriod,
		connectionTimeout: staleConnectionTimeout,
	}
	cleaner.start()

	go func() {
		statsPeriod := time.Second * time.Duration(*statsevery)
		ticker := time.NewTicker(statsPeriod)

		for true {
			<-ticker.C

			// Since the start
			errorsMapMutex.Lock()
			errorMapLen := len(errorsMap)
			errorsSummery := fmt.Sprintf("%v", errorsMap)
			errorsMapMutex.Unlock()
			fmt.Printf("Processed %v packets (%v bytes) in %v (errors: %v, errTypes:%v)\nErrors Summary: %s\n",
				count,
				bytes,
				time.Since(start),
				nErrors,
				errorMapLen,
				errorsSummery,
			)

			// At this moment
			memStats := runtime.MemStats{}
			runtime.ReadMemStats(&memStats)
			fmt.Printf(
				"mem: %d, goroutines: %d, unmatched messages: %d\n",
				memStats.HeapAlloc,
				runtime.NumGoroutine(),
				reqResMatcher.openMessagesMap.Count(),
			)

			// Since the last print
			cleanStats := cleaner.dumpStats()
			appStats := statsTracker.dumpStats()
			fmt.Printf(
				"flushed connections %d, closed connections: %d, deleted messages: %d, matched messages: %d\n",
				cleanStats.flushed,
				cleanStats.closed,
				cleanStats.deleted,
				appStats.matchedMessages,
			)
		}
	}()

	for packet := range source.Packets() {
		count++
		Debug("PACKET #%d\n", count)
		data := packet.Data()
		bytes += int64(len(data))
		if *hexdumppkt {
			Debug("Packet content (%d/0x%x)\n%s\n", len(data), len(data), hex.Dump(data))
		}

		// defrag the IPv4 packet if required
		if !*nodefrag {
			ip4Layer := packet.Layer(layers.LayerTypeIPv4)
			if ip4Layer == nil {
				continue
			}
			ip4 := ip4Layer.(*layers.IPv4)
			l := ip4.Length
			newip4, err := defragger.DefragIPv4(ip4)
			if err != nil {
				log.Fatalln("Error while de-fragmenting", err)
			} else if newip4 == nil {
				Debug("Fragment...\n")
				continue // packet fragment, we don't have whole packet yet.
			}
			if newip4.Length != l {
				stats.ipdefrag++
				Debug("Decoding re-assembled packet: %s\n", newip4.NextLayerType())
				pb, ok := packet.(gopacket.PacketBuilder)
				if !ok {
					panic("Not a PacketBuilder")
				}
				nextDecoder := newip4.NextLayerType()
				nextDecoder.Decode(newip4.Payload, pb)
			}
		}

		tcp := packet.Layer(layers.LayerTypeTCP)
		if tcp != nil {
			tcp := tcp.(*layers.TCP)
			if *checksum {
				err := tcp.SetNetworkLayerForChecksum(packet.NetworkLayer())
				if err != nil {
					log.Fatalf("Failed to set network layer for checksum: %s\n", err)
				}
			}
			c := Context{
				CaptureInfo: packet.Metadata().CaptureInfo,
			}
			stats.totalsz += len(tcp.Payload)
			//fmt.Println(packet.NetworkLayer().NetworkFlow().Src(), ":", tcp.SrcPort, " -> ", packet.NetworkLayer().NetworkFlow().Dst(), ":", tcp.DstPort)
			assemblerMutex.Lock()
			assembler.AssembleWithContext(packet.NetworkLayer().NetworkFlow(), tcp, &c)
			assemblerMutex.Unlock()
		}

		done := *maxcount > 0 && count >= *maxcount
		if done {
			errorsMapMutex.Lock()
			errorMapLen := len(errorsMap)
			errorsMapMutex.Unlock()
			fmt.Fprintf(os.Stderr, "Processed %v packets (%v bytes) in %v (errors: %v, errTypes:%v)\n", count, bytes, time.Since(start), nErrors, errorMapLen)
		}
		select {
		case <-signalChan:
			fmt.Fprintf(os.Stderr, "\nCaught SIGINT: aborting\n")
			done = true
		default:
			// NOP: continue
		}
		if done {
			break
		}
	}

	assemblerMutex.Lock()
	closed := assembler.FlushAll()
	assemblerMutex.Unlock()
	Debug("Final flush: %d closed", closed)
	if outputLevel >= 2 {
		streamPool.Dump()
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		f.Close()
	}

	streamFactory.WaitGoRoutines()
	assemblerMutex.Lock()
	Debug("%s\n", assembler.Dump())
	assemblerMutex.Unlock()
	if !*nodefrag {
		fmt.Printf("IPdefrag:\t\t%d\n", stats.ipdefrag)
	}
	fmt.Printf("TCP stats:\n")
	fmt.Printf(" missed bytes:\t\t%d\n", stats.missedBytes)
	fmt.Printf(" total packets:\t\t%d\n", stats.pkt)
	fmt.Printf(" rejected FSM:\t\t%d\n", stats.rejectFsm)
	fmt.Printf(" rejected Options:\t%d\n", stats.rejectOpt)
	fmt.Printf(" reassembled bytes:\t%d\n", stats.sz)
	fmt.Printf(" total TCP bytes:\t%d\n", stats.totalsz)
	fmt.Printf(" conn rejected FSM:\t%d\n", stats.rejectConnFsm)
	fmt.Printf(" reassembled chunks:\t%d\n", stats.reassembled)
	fmt.Printf(" out-of-order packets:\t%d\n", stats.outOfOrderPackets)
	fmt.Printf(" out-of-order bytes:\t%d\n", stats.outOfOrderBytes)
	fmt.Printf(" biggest-chunk packets:\t%d\n", stats.biggestChunkPackets)
	fmt.Printf(" biggest-chunk bytes:\t%d\n", stats.biggestChunkBytes)
	fmt.Printf(" overlap packets:\t%d\n", stats.overlapPackets)
	fmt.Printf(" overlap bytes:\t\t%d\n", stats.overlapBytes)
	fmt.Printf("Errors: %d\n", nErrors)
	for e := range errorsMap {
		fmt.Printf(" %s:\t\t%d\n", e, errorsMap[e])
	}
}
