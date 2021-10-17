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
	"io"
	"os"
	"os/signal"
	"runtime"
	_debug "runtime/debug"
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
	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/tap/api"
)

const cleanPeriod = time.Second * 10

var remoteOnlyOutboundPorts = []int{80, 443}

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
var hexdumppkt = flag.Bool("dumppkt", false, "Dump packet as hex")

// capture
var iface = flag.String("i", "en0", "Interface to read packets from")
var fname = flag.String("r", "", "Filename to read from, overrides -i")
var snaplen = flag.Int("s", 65536, "Snap length (number of bytes max to read per packet")
var tstype = flag.String("timestamp_type", "", "Type of timestamps to use")
var promisc = flag.Bool("promisc", true, "Set promiscuous mode")
var staleTimeoutSeconds = flag.Int("staletimout", 120, "Max time in seconds to keep connections which don't transmit data")

var memprofile = flag.String("memprofile", "", "Write memory profile")

var appStats = api.AppStats{}

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
var ownIps []string                               // global
var hostMode bool                                 // global
var extensions []*api.Extension                   // global
var filteringOptions *api.TrafficFilteringOptions // global

const baseStreamChannelTimeoutMs int = 5000 * 100

/* minOutputLevel: Error will be printed only if outputLevel is above this value
 * t:              key for errorsMap (counting errors)
 * s, a:           arguments logger.Log.Infof
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
		logger.Log.Errorf(formatStr, a...)
	}
}
func Error(t string, s string, a ...interface{}) {
	logError(0, t, s, a...)
}
func SilentError(t string, s string, a ...interface{}) {
	logError(2, t, s, a...)
}
func Debug(s string, a ...interface{}) {
	logger.Log.Debugf(s, a...)
}
func Trace(s string, a ...interface{}) {
	logger.Log.Infof(s, a...)
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

func GetStats() api.AppStats {
	return appStats
}

func (c *Context) GetCaptureInfo() gopacket.CaptureInfo {
	return c.CaptureInfo
}

func StartPassiveTapper(opts *TapOpts, outputItems chan *api.OutputChannelItem, extensionsRef []*api.Extension, options *api.TrafficFilteringOptions) {
	hostMode = opts.HostMode
	extensions = extensionsRef
	filteringOptions = options

	if GetMemoryProfilingEnabled() {
		startMemoryProfiler()
	}

	go startPassiveTapper(outputItems)
}

func startMemoryProfiler() {
	dumpPath := "/app/pprof"
	envDumpPath := os.Getenv(MemoryProfilingDumpPath)
	if envDumpPath != "" {
		dumpPath = envDumpPath
	}
	timeInterval := 60
	envTimeInterval := os.Getenv(MemoryProfilingTimeIntervalSeconds)
	if envTimeInterval != "" {
		if i, err := strconv.Atoi(envTimeInterval); err == nil {
			timeInterval = i
		}
	}

	logger.Log.Info("Profiling is on, results will be written to %s", dumpPath)
	go func() {
		if _, err := os.Stat(dumpPath); os.IsNotExist(err) {
			if err := os.Mkdir(dumpPath, 0777); err != nil {
				logger.Log.Fatal("could not create directory for profile: ", err)
			}
		}

		for true {
			t := time.Now()

			filename := fmt.Sprintf("%s/%s__mem.prof", dumpPath, t.Format("15_04_05"))

			logger.Log.Infof("Writing memory profile to %s\n", filename)

			f, err := os.Create(filename)
			if err != nil {
				logger.Log.Fatal("could not create memory profile: ", err)
			}
			runtime.GC() // get up-to-date statistics
			if err := pprof.WriteHeapProfile(f); err != nil {
				logger.Log.Fatal("could not write memory profile: ", err)
			}
			_ = f.Close()
			time.Sleep(time.Second * time.Duration(timeInterval))
		}
	}()
}

func closeTimedoutTcpStreamChannels() {
	TcpStreamChannelTimeoutMs := GetTcpChannelTimeoutMs()
	for {
		time.Sleep(10 * time.Millisecond)
		_debug.FreeOSMemory()
		streams.Range(func(key interface{}, value interface{}) bool {
			streamWrapper := value.(*tcpStreamWrapper)
			stream := streamWrapper.stream
			if stream.superIdentifier.Protocol == nil {
				if !stream.isClosed && time.Now().After(streamWrapper.createdAt.Add(TcpStreamChannelTimeoutMs)) {
					stream.Close()
					appStats.IncDroppedTcpStreams()
					logger.Log.Debugf("Dropped an unidentified TCP stream because of timeout. Total dropped: %d Total Goroutines: %d Timeout (ms): %d\n", appStats.DroppedTcpStreams, runtime.NumGoroutine(), TcpStreamChannelTimeoutMs/1000000)
				}
			} else {
				if !stream.superIdentifier.IsClosedOthers {
					for i := range stream.clients {
						reader := &stream.clients[i]
						if reader.extension.Protocol != stream.superIdentifier.Protocol {
							reader.Close()
						}
					}
					for i := range stream.servers {
						reader := &stream.servers[i]
						if reader.extension.Protocol != stream.superIdentifier.Protocol {
							reader.Close()
						}
					}
					stream.superIdentifier.IsClosedOthers = true
				}
			}
			return true
		})
	}
}

func startPassiveTapper(outputItems chan *api.OutputChannelItem) {
	go closeTimedoutTcpStreamChannels()

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
		logger.Log.Info("Failed to get self IP addresses")
		logger.Log.Errorf("Getting-Self-Address", "Error getting self ip address: %s (%v,%+v)", err, err, err)
		ownIps = make([]string, 0)
	} else {
		ownIps = localhostIPs
	}

	var handle *pcap.Handle
	var err error
	if *fname != "" {
		if handle, err = pcap.OpenOffline(*fname); err != nil {
			logger.Log.Fatalf("PCAP OpenOffline error: %v", err)
		}
	} else {
		// This is a little complicated because we want to allow all possible options
		// for creating the packet capture handle... instead of all this you can
		// just call pcap.OpenLive if you want a simple handle.
		inactive, err := pcap.NewInactiveHandle(*iface)
		if err != nil {
			logger.Log.Fatalf("could not create: %v", err)
		}
		defer inactive.CleanUp()
		if err = inactive.SetSnapLen(*snaplen); err != nil {
			logger.Log.Fatalf("could not set snap length: %v", err)
		} else if err = inactive.SetPromisc(*promisc); err != nil {
			logger.Log.Fatalf("could not set promisc mode: %v", err)
		} else if err = inactive.SetTimeout(time.Second); err != nil {
			logger.Log.Fatalf("could not set timeout: %v", err)
		}
		if *tstype != "" {
			if t, err := pcap.TimestampSourceFromString(*tstype); err != nil {
				logger.Log.Fatalf("Supported timestamp types: %v", inactive.SupportedTimestamps())
			} else if err := inactive.SetTimestampSource(t); err != nil {
				logger.Log.Fatalf("Supported timestamp types: %v", inactive.SupportedTimestamps())
			}
		}
		if handle, err = inactive.Activate(); err != nil {
			logger.Log.Fatalf("PCAP Activate error: %v", err)
		}
		defer handle.Close()
	}
	if len(flag.Args()) > 0 {
		bpffilter := strings.Join(flag.Args(), " ")
		logger.Log.Infof("Using BPF filter %q", bpffilter)
		if err = handle.SetBPFFilter(bpffilter); err != nil {
			logger.Log.Fatalf("BPF filter error: %v", err)
		}
	}

	var dec gopacket.Decoder
	var ok bool
	decoderName := *decoder
	if decoderName == "" {
		decoderName = fmt.Sprintf("%s", handle.LinkType())
	}
	if dec, ok = gopacket.DecodersByLayerName[decoderName]; !ok {
		logger.Log.Fatal("No decoder named", decoderName)
	}
	source := gopacket.NewPacketSource(handle, dec)
	source.Lazy = *lazy
	source.NoCopy = true
	logger.Log.Info("Starting to read packets")
	appStats.SetStartTime(time.Now())
	defragger := ip4defrag.NewIPv4Defragmenter()

	var emitter api.Emitter = &api.Emitting{
		AppStats:      &appStats,
		OutputChannel: outputItems,
	}

	streamFactory := &tcpStreamFactory{
		Emitter: emitter,
	}
	streamPool := reassembly.NewStreamPool(streamFactory)
	assembler := reassembly.NewAssembler(streamPool)

	maxBufferedPagesTotal := GetMaxBufferedPagesPerConnection()
	maxBufferedPagesPerConnection := GetMaxBufferedPagesTotal()
	logger.Log.Infof("Assembler options: maxBufferedPagesTotal=%d, maxBufferedPagesPerConnection=%d", maxBufferedPagesTotal, maxBufferedPagesPerConnection)
	assembler.AssemblerOptions.MaxBufferedPagesTotal = maxBufferedPagesTotal
	assembler.AssemblerOptions.MaxBufferedPagesPerConnection = maxBufferedPagesPerConnection

	var assemblerMutex sync.Mutex

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	staleConnectionTimeout := time.Second * time.Duration(*staleTimeoutSeconds)
	cleaner := Cleaner{
		assembler:         assembler,
		assemblerMutex:    &assemblerMutex,
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
			logger.Log.Infof("%v (errors: %v, errTypes:%v) - Errors Summary: %s",
				time.Since(appStats.StartTime),
				nErrors,
				errorMapLen,
				errorsSummery,
			)

			// At this moment
			memStats := runtime.MemStats{}
			runtime.ReadMemStats(&memStats)
			logger.Log.Infof(
				"mem: %d, goroutines: %d",
				memStats.HeapAlloc,
				runtime.NumGoroutine(),
			)

			// Since the last print
			cleanStats := cleaner.dumpStats()
			logger.Log.Infof(
				"cleaner - flushed connections: %d, closed connections: %d, deleted messages: %d",
				cleanStats.flushed,
				cleanStats.closed,
				cleanStats.deleted,
			)
			currentAppStats := appStats.DumpStats()
			appStatsJSON, _ := json.Marshal(currentAppStats)
			logger.Log.Infof("app stats - %v", string(appStatsJSON))
		}
	}()

	if GetMemoryProfilingEnabled() {
		startMemoryProfiler()
	}

	for {
		packet, err := source.NextPacket()
		if err == io.EOF {
			break
		} else if err != nil {
			logger.Log.Debugf("Error: %v", err)
			continue
		}
		packetsCount := appStats.IncPacketsCount()
		logger.Log.Debugf("PACKET #%d", packetsCount)
		data := packet.Data()
		appStats.UpdateProcessedBytes(uint64(len(data)))
		if *hexdumppkt {
			logger.Log.Debugf("Packet content (%d/0x%x) - %s", len(data), len(data), hex.Dump(data))
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
				logger.Log.Fatal("Error while de-fragmenting", err)
			} else if newip4 == nil {
				logger.Log.Debugf("Fragment...")
				continue // packet fragment, we don't have whole packet yet.
			}
			if newip4.Length != l {
				stats.ipdefrag++
				logger.Log.Debugf("Decoding re-assembled packet: %s", newip4.NextLayerType())
				pb, ok := packet.(gopacket.PacketBuilder)
				if !ok {
					logger.Log.Panic("Not a PacketBuilder")
				}
				nextDecoder := newip4.NextLayerType()
				_ = nextDecoder.Decode(newip4.Payload, pb)
			}
		}

		tcp := packet.Layer(layers.LayerTypeTCP)
		if tcp != nil {
			appStats.IncTcpPacketsCount()
			tcp := tcp.(*layers.TCP)
			if *checksum {
				err := tcp.SetNetworkLayerForChecksum(packet.NetworkLayer())
				if err != nil {
					logger.Log.Fatalf("Failed to set network layer for checksum: %s\n", err)
				}
			}
			c := Context{
				CaptureInfo: packet.Metadata().CaptureInfo,
			}
			stats.totalsz += len(tcp.Payload)
			logger.Log.Debugf("%s : %v -> %s : %v", packet.NetworkLayer().NetworkFlow().Src(), tcp.SrcPort, packet.NetworkLayer().NetworkFlow().Dst(), tcp.DstPort)
			assemblerMutex.Lock()
			assembler.AssembleWithContext(packet.NetworkLayer().NetworkFlow(), tcp, &c)
			assemblerMutex.Unlock()
		}

		done := *maxcount > 0 && int64(appStats.PacketsCount) >= *maxcount
		if done {
			errorsMapMutex.Lock()
			errorMapLen := len(errorsMap)
			errorsMapMutex.Unlock()
			logger.Log.Infof("Processed %v packets (%v bytes) in %v (errors: %v, errTypes:%v)",
				appStats.PacketsCount,
				appStats.ProcessedBytes,
				time.Since(appStats.StartTime),
				nErrors,
				errorMapLen)
		}
		select {
		case <-signalChan:
			logger.Log.Infof("Caught SIGINT: aborting")
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
	logger.Log.Debugf("Final flush: %d closed", closed)
	if outputLevel >= 2 {
		streamPool.Dump()
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			logger.Log.Fatal(err)
		}
		_ = pprof.WriteHeapProfile(f)
		_ = f.Close()
	}

	streamFactory.WaitGoRoutines()
	assemblerMutex.Lock()
	logger.Log.Debugf("%s", assembler.Dump())
	assemblerMutex.Unlock()
	if !*nodefrag {
		logger.Log.Infof("IPdefrag:\t\t%d", stats.ipdefrag)
	}
	logger.Log.Infof("TCP stats:")
	logger.Log.Infof(" missed bytes:\t\t%d", stats.missedBytes)
	logger.Log.Infof(" total packets:\t\t%d", stats.pkt)
	logger.Log.Infof(" rejected FSM:\t\t%d", stats.rejectFsm)
	logger.Log.Infof(" rejected Options:\t%d", stats.rejectOpt)
	logger.Log.Infof(" reassembled bytes:\t%d", stats.sz)
	logger.Log.Infof(" total TCP bytes:\t%d", stats.totalsz)
	logger.Log.Infof(" conn rejected FSM:\t%d", stats.rejectConnFsm)
	logger.Log.Infof(" reassembled chunks:\t%d", stats.reassembled)
	logger.Log.Infof(" out-of-order packets:\t%d", stats.outOfOrderPackets)
	logger.Log.Infof(" out-of-order bytes:\t%d", stats.outOfOrderBytes)
	logger.Log.Infof(" biggest-chunk packets:\t%d", stats.biggestChunkPackets)
	logger.Log.Infof(" biggest-chunk bytes:\t%d", stats.biggestChunkBytes)
	logger.Log.Infof(" overlap packets:\t%d", stats.overlapPackets)
	logger.Log.Infof(" overlap bytes:\t\t%d", stats.overlapBytes)
	logger.Log.Infof("Errors: %d", nErrors)
	for e := range errorsMap {
		logger.Log.Infof(" %s:\t\t%d", e, errorsMap[e])
	}
	logger.Log.Infof("AppStats: %v", GetStats())
}
