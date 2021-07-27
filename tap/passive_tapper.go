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

	"github.com/romana/rlog"

	"github.com/google/gopacket"
	"github.com/google/gopacket/examples/util"
	"github.com/google/gopacket/ip4defrag"
	"github.com/google/gopacket/layers" // pulls in all layers decoders
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/reassembly"
)

const AppPortsEnvVar = "APP_PORTS"
const maxHTTP2DataLenEnvVar = "HTTP2_DATA_SIZE_LIMIT"
const maxHTTP2DataLenDefault = 1 * 1024 * 1024 // 1MB
const cleanPeriod = time.Second * 10

var remoteOnlyOutboundPorts = []int{80, 443}

func parseAppPorts(appPortsList string) []int {
	ports := make([]int, 0)
	for _, portStr := range strings.Split(appPortsList, ",") {
		parsedInt, parseError := strconv.Atoi(portStr)
		if parseError != nil {
			log.Printf("Provided app port %v is not a valid number!", portStr)
		} else {
			ports = append(ports, parsedInt)
		}
	}
	return ports
}

var maxcount = flag.Int("c", -1, "Only grab this many packets, then exit")
var decoder = flag.String("decoder", "", "Name of the decoder to use (default: guess from capture)")
var statsevery = flag.Int("stats", 60, "Output statistics every N seconds")
var lazy = flag.Bool("lazy", false, "If true, do lazy decoding")
var nodefrag = flag.Bool("nodefrag", false, "If true, do not do IPv4 defrag")
var checksum = flag.Bool("checksum", false, "Check TCP checksum")                                                      // global
var nooptcheck = flag.Bool("nooptcheck", true, "Do not check TCP options (useful to ignore MSS on captures with TSO)") // global
var ignorefsmerr = flag.Bool("ignorefsmerr", true, "Ignore TCP FSM errors")                                            // global
var allowmissinginit = flag.Bool("allowmissinginit", true, "Support streams without SYN/SYN+ACK/ACK sequence")         // global
var verbose = flag.Bool("verbose", false, "Be verbose")
var debug = flag.Bool("debug", false, "Display debug information")
var quiet = flag.Bool("quiet", false, "Be quiet regarding errors")

// http
var nohttp = flag.Bool("nohttp", false, "Disable HTTP parsing")
var output = flag.String("output", "", "Path to create file for HTTP 200 OK responses")
var writeincomplete = flag.Bool("writeincomplete", false, "Write incomplete response")

var hexdump = flag.Bool("dump", false, "Dump HTTP request/response as hex") // global
var hexdumppkt = flag.Bool("dumppkt", false, "Dump packet as hex")

// capture
var iface = flag.String("i", "eth0", "Interface to read packets from")
var fname = flag.String("r", "", "Filename to read from, overrides -i")
var snaplen = flag.Int("s", 65536, "Snap length (number of bytes max to read per packet")
var tstype = flag.String("timestamp_type", "", "Type of timestamps to use")
var promisc = flag.Bool("promisc", true, "Set promiscuous mode")
var anydirection = flag.Bool("anydirection", false, "Capture http requests to other hosts")
var staleTimeoutSeconds = flag.Int("staletimout", 120, "Max time in seconds to keep connections which don't transmit data")

var memprofile = flag.String("memprofile", "", "Write memory profile")

// output
var dumpToHar = flag.Bool("hardump", false, "Dump traffic to har files")
var HarOutputDir = flag.String("hardir", "", "Directory in which to store output har files")
var harEntriesPerFile = flag.Int("harentriesperfile", 200, "Number of max number of har entries to store in each file")

var reqResMatcher = createResponseRequestMatcher() // global
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

type TapOpts struct {
	HostMode bool
}

var outputLevel int
var errorsMap map[string]uint
var errorsMapMutex sync.Mutex
var nErrors uint
var ownIps []string // global
var hostMode bool   // global

/* minOutputLevel: Error will be printed only if outputLevel is above this value
 * t:              key for errorsMap (counting errors)
 * s, a:           arguments log.Printf
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
		rlog.Errorf(formatStr, a...)
	}
}
func Error(t string, s string, a ...interface{}) {
	logError(0, t, s, a...)
}
func SilentError(t string, s string, a ...interface{}) {
	logError(2, t, s, a...)
}
func Debug(s string, a ...interface{}) {
	rlog.Debugf(s, a...)
}
func Trace(s string, a ...interface{}) {
	rlog.Tracef(1, s, a...)
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

// Context
// The assembler context
type Context struct {
	CaptureInfo gopacket.CaptureInfo
}

func (c *Context) GetCaptureInfo() gopacket.CaptureInfo {
	return c.CaptureInfo
}

func StartPassiveTapper(opts *TapOpts) (<-chan *OutputChannelItem, <-chan *OutboundLink) {
	hostMode = opts.HostMode

	var harWriter *HarWriter
	if *dumpToHar {
		harWriter = NewHarWriter(*HarOutputDir, *harEntriesPerFile)
	}
	outboundLinkWriter := NewOutboundLinkWriter()

	go startPassiveTapper(harWriter, outboundLinkWriter)

	if harWriter != nil {
		return harWriter.OutChan, outboundLinkWriter.OutChan
	}

	return nil, outboundLinkWriter.OutChan
}

func startPassiveTapper(harWriter *HarWriter, outboundLinkWriter *OutboundLinkWriter) {
	os.Setenv("RLOG_LOG_LEVEL", "DEBUG") // TODO: revert this `$ export RLOG_LOG_LEVEL=DEBUG` not working for some reason
	rlog.UpdateEnv()

	log.SetFlags(log.LstdFlags | log.LUTC | log.Lshortfile)

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
		rlog.Info("Failed to get self IP addresses")
		rlog.Errorf("Getting-Self-Address", "Error getting self ip address: %s (%v,%+v)", err, err, err)
		ownIps = make([]string, 0)
	} else {
		ownIps = localhostIPs
	}

	appPortsStr := os.Getenv(AppPortsEnvVar)
	var appPorts []int
	if appPortsStr == "" {
		rlog.Info("Received empty/no APP_PORTS env var! only listening to http on port 80!")
		appPorts = make([]int, 0)
	} else {
		appPorts = parseAppPorts(appPortsStr)
	}
	SetFilterPorts(appPorts)
	envVal := os.Getenv(maxHTTP2DataLenEnvVar)
	if envVal == "" {
		rlog.Infof("Received empty/no HTTP2_DATA_SIZE_LIMIT env var! falling back to %v", maxHTTP2DataLenDefault)
		maxHTTP2DataLen = maxHTTP2DataLenDefault
	} else {
		if convertedInt, err := strconv.Atoi(envVal); err != nil {
			rlog.Infof("Received invalid HTTP2_DATA_SIZE_LIMIT env var! falling back to %v", maxHTTP2DataLenDefault)
			maxHTTP2DataLen = maxHTTP2DataLenDefault
		} else {
			rlog.Infof("Received HTTP2_DATA_SIZE_LIMIT env var: %v", maxHTTP2DataLenDefault)
			maxHTTP2DataLen = convertedInt
		}
	}

	log.Printf("App Ports: %v", gSettings.filterPorts)

	var handle *pcap.Handle
	var err error
	if *fname != "" {
		if handle, err = pcap.OpenOffline(*fname); err != nil {
			log.Fatalf("PCAP OpenOffline error: %v", err)
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
			log.Fatalf("PCAP Activate error: %v", err)
		}
		defer handle.Close()
	}
	if len(flag.Args()) > 0 {
		bpffilter := strings.Join(flag.Args(), " ")
		rlog.Infof("Using BPF filter %q", bpffilter)
		if err = handle.SetBPFFilter(bpffilter); err != nil {
			log.Fatalf("BPF filter error: %v", err)
		}
	}

	if *dumpToHar {
		harWriter.Start()
		defer harWriter.Stop()
	}
	defer outboundLinkWriter.Stop()

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
	rlog.Info("Starting to read packets")
	count := 0
	bytes := int64(0)
	start := time.Now()
	defragger := ip4defrag.NewIPv4Defragmenter()

	streamFactory := &tcpStreamFactory{
		doHTTP:             !*nohttp,
		harWriter:          harWriter,
		outbountLinkWriter: outboundLinkWriter,
	}
	streamPool := reassembly.NewStreamPool(streamFactory)
	assembler := reassembly.NewAssembler(streamPool)
	var assemblerMutex sync.Mutex

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	staleConnectionTimeout := time.Second * time.Duration(*staleTimeoutSeconds)
	cleaner := Cleaner{
		assembler:         assembler,
		assemblerMutex:    &assemblerMutex,
		matcher:           &reqResMatcher,
		cleanPeriod:       cleanPeriod,
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
			log.Printf("Processed %v packets (%v bytes) in %v (errors: %v, errTypes:%v) - Errors Summary: %s",
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
			log.Printf(
				"mem: %d, goroutines: %d, unmatched messages: %d",
				memStats.HeapAlloc,
				runtime.NumGoroutine(),
				reqResMatcher.openMessagesMap.Count(),
			)

			// Since the last print
			cleanStats := cleaner.dumpStats()
			appStats := statsTracker.dumpStats()
			log.Printf(
				"flushed connections %d, closed connections: %d, deleted messages: %d, matched messages: %d",
				cleanStats.flushed,
				cleanStats.closed,
				cleanStats.deleted,
				appStats.matchedMessages,
			)
		}
	}()

	for packet := range source.Packets() {
		count++
		rlog.Debugf("PACKET #%d", count)
		data := packet.Data()
		bytes += int64(len(data))
		if *hexdumppkt {
			rlog.Debugf("Packet content (%d/0x%x) - %s", len(data), len(data), hex.Dump(data))
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
				rlog.Debugf("Fragment...")
				continue // packet fragment, we don't have whole packet yet.
			}
			if newip4.Length != l {
				stats.ipdefrag++
				rlog.Debugf("Decoding re-assembled packet: %s", newip4.NextLayerType())
				pb, ok := packet.(gopacket.PacketBuilder)
				if !ok {
					log.Panic("Not a PacketBuilder")
				}
				nextDecoder := newip4.NextLayerType()
				_ = nextDecoder.Decode(newip4.Payload, pb)
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
			rlog.Debugf("%s : %v -> %s : %v", packet.NetworkLayer().NetworkFlow().Src(), tcp.SrcPort, packet.NetworkLayer().NetworkFlow().Dst(), tcp.DstPort)
			assemblerMutex.Lock()
			assembler.AssembleWithContext(packet.NetworkLayer().NetworkFlow(), tcp, &c)
			assemblerMutex.Unlock()
		}

		done := *maxcount > 0 && count >= *maxcount
		if done {
			errorsMapMutex.Lock()
			errorMapLen := len(errorsMap)
			errorsMapMutex.Unlock()
			log.Printf("Processed %v packets (%v bytes) in %v (errors: %v, errTypes:%v)", count, bytes, time.Since(start), nErrors, errorMapLen)
		}
		select {
		case <-signalChan:
			log.Printf("Caught SIGINT: aborting")
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
	rlog.Debugf("Final flush: %d closed", closed)
	if outputLevel >= 2 {
		streamPool.Dump()
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		_ = pprof.WriteHeapProfile(f)
		_ = f.Close()
	}

	streamFactory.WaitGoRoutines()
	assemblerMutex.Lock()
	rlog.Debugf("%s", assembler.Dump())
	assemblerMutex.Unlock()
	if !*nodefrag {
		log.Printf("IPdefrag:\t\t%d", stats.ipdefrag)
	}
	log.Printf("TCP stats:")
	log.Printf(" missed bytes:\t\t%d", stats.missedBytes)
	log.Printf(" total packets:\t\t%d", stats.pkt)
	log.Printf(" rejected FSM:\t\t%d", stats.rejectFsm)
	log.Printf(" rejected Options:\t%d", stats.rejectOpt)
	log.Printf(" reassembled bytes:\t%d", stats.sz)
	log.Printf(" total TCP bytes:\t%d", stats.totalsz)
	log.Printf(" conn rejected FSM:\t%d", stats.rejectConnFsm)
	log.Printf(" reassembled chunks:\t%d", stats.reassembled)
	log.Printf(" out-of-order packets:\t%d", stats.outOfOrderPackets)
	log.Printf(" out-of-order bytes:\t%d", stats.outOfOrderBytes)
	log.Printf(" biggest-chunk packets:\t%d", stats.biggestChunkPackets)
	log.Printf(" biggest-chunk bytes:\t%d", stats.biggestChunkBytes)
	log.Printf(" overlap packets:\t%d", stats.overlapPackets)
	log.Printf(" overlap bytes:\t\t%d", stats.overlapBytes)
	log.Printf("Errors: %d", nErrors)
	for e := range errorsMap {
		log.Printf(" %s:\t\t%d", e, errorsMap[e])
	}
}
