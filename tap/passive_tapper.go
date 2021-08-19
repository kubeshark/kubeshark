// Copyright 2012 Google, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

// The pcapdump binary implements a tcpdump-like command line tool with gopacket
// using pcap as a backend data collection mechanism.
package tap

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"plugin"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/romana/rlog"

	"github.com/google/gopacket"
	"github.com/google/gopacket/examples/util"
	"github.com/google/gopacket/layers" // pulls in all layers decoders
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
	"github.com/up9inc/mizu/tap/api"
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

var maxcount = flag.Int64("c", -1, "Only grab this many packets, then exit")
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
var iface = flag.String("i", "en0", "Interface to read packets from")
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
var filter = flag.String("f", "tcp", "BPF filter for pcap")

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
var ownIps []string             // global
var hostMode bool               // global
var extensions []*api.Extension // global
var allOutboundPorts []string   // global
var allInboundPorts []string    // global

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

func GetStats() AppStats {
	return statsTracker.appStats
}

func (c *Context) GetCaptureInfo() gopacket.CaptureInfo {
	return c.CaptureInfo
}

func StartPassiveTapper(opts *TapOpts, outputItems chan *api.OutputChannelItem) {
	hostMode = opts.HostMode

	if GetMemoryProfilingEnabled() {
		startMemoryProfiler()
	}

	go startPassiveTapper(outputItems)
}

func startMemoryProfiler() {
	dirname := "/app/pprof"
	rlog.Info("Profiling is on, results will be written to %s", dirname)
	go func() {
		if _, err := os.Stat(dirname); os.IsNotExist(err) {
			if err := os.Mkdir(dirname, 0777); err != nil {
				log.Fatal("could not create directory for profile: ", err)
			}
		}

		for true {
			t := time.Now()

			filename := fmt.Sprintf("%s/%s__mem.prof", dirname, t.Format("15_04_05"))

			rlog.Info("Writing memory profile to %s\n", filename)

			f, err := os.Create(filename)
			if err != nil {
				log.Fatal("could not create memory profile: ", err)
			}
			runtime.GC() // get up-to-date statistics
			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Fatal("could not write memory profile: ", err)
			}
			_ = f.Close()
			time.Sleep(time.Minute)
		}
	}()
}

func MergeUnique(slice []string, merge []string) []string {
	for _, i := range merge {
		add := true
		for _, ele := range slice {
			if ele == i {
				add = false
			}
		}
		if add {
			slice = append(slice, i)
		}
	}
	return slice
}

func loadExtensions() {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	extensionsDir := path.Join(dir, "./extensions/")

	files, err := ioutil.ReadDir(extensionsDir)
	if err != nil {
		log.Fatal(err)
	}
	extensions = make([]*api.Extension, len(files))
	for i, file := range files {
		filename := file.Name()
		log.Printf("Loading extension: %s\n", filename)
		extension := &api.Extension{
			Path: path.Join(extensionsDir, filename),
		}
		plug, _ := plugin.Open(extension.Path)
		extension.Plug = plug
		symDissector, _ := plug.Lookup("Dissector")

		var dissector api.Dissector
		dissector, _ = symDissector.(api.Dissector)
		dissector.Register(extension)
		extension.Dissector = dissector
		log.Printf("Extension Properties: %+v\n", extension)
		extensions[i] = extension
		allOutboundPorts = MergeUnique(allOutboundPorts, extension.OutboundPorts)
		allInboundPorts = MergeUnique(allInboundPorts, extension.InboundPorts)
	}
	log.Printf("allOutboundPorts: %v\n", allOutboundPorts)
	log.Printf("allInboundPorts: %v\n", allInboundPorts)
}

func startPassiveTapper(outputItems chan *api.OutputChannelItem) {
	loadExtensions()

	log.SetFlags(log.LstdFlags | log.LUTC | log.Lshortfile)

	defer util.Run()()
	var handle *pcap.Handle
	var err error

	if localhostIPs, err := getLocalhostIPs(); err != nil {
		// TODO: think this over
		rlog.Info("Failed to get self IP addresses")
		rlog.Errorf("Getting-Self-Address", "Error getting self ip address: %s (%v,%+v)", err, err, err)
		ownIps = make([]string, 0)
	} else {
		ownIps = localhostIPs
	}

	// Set up pcap packet capture
	if *fname != "" {
		log.Printf("Reading from pcap dump %q", *fname)
		handle, err = pcap.OpenOffline(*fname)
	} else {
		log.Printf("Starting capture on interface %q", *iface)
		handle, err = pcap.OpenLive(*iface, int32(*snaplen), true, pcap.BlockForever)
	}
	if err != nil {
		log.Fatal(err)
	}

	if err := handle.SetBPFFilter(*filter); err != nil {
		log.Fatal(err)
	}

	var emitter api.Emitter = &api.Emitting{
		OutputChannel: outputItems,
	}

	// Set up assembly
	streamFactory := &tcpStreamFactory{
		Emitter: emitter,
	}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)

	log.Println("reading in packets")
	// Read in packets, pass to assembler.
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packets := packetSource.Packets()
	ticker := time.Tick(time.Minute)
	for {
		select {
		case packet := <-packets:
			// A nil packet indicates the end of a pcap file.
			if packet == nil {
				return
			}
			if *verbose {
				log.Println(packet)
			}
			if packet.NetworkLayer() == nil || packet.TransportLayer() == nil || packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
				log.Println("Unusable packet")
				continue
			}
			tcp := packet.TransportLayer().(*layers.TCP)
			assembler.AssembleWithTimestamp(packet.NetworkLayer().NetworkFlow(), tcp, packet.Metadata().Timestamp)

		case <-ticker:
			// Every minute, flush connections that haven't seen activity in the past 2 minutes.
			assembler.FlushOlderThan(time.Now().Add(time.Minute * -2))
		}
	}
}
