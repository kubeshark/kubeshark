// Copyright 2012 Google, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

// The pcapdump binary implements a tcpdump-like command line tool with gopacket
// using pcap as a backend data collection mechanism.
package tap

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/examples/util" // pulls in all layers decoders
	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/tap/api"
)

const cleanPeriod = time.Second * 10

//lint:ignore U1000 will be used in the future
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
var tapErrors *errorsMap

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

var hostMode bool                                 // global
var extensions []*api.Extension                   // global
var filteringOptions *api.TrafficFilteringOptions // global

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

		for {
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

func printPeriodicStats(cleaner *Cleaner) {
	statsPeriod := time.Second * time.Duration(*statsevery)
	ticker := time.NewTicker(statsPeriod)

	for {
		<-ticker.C

		// Since the start
		errorMapLen, errorsSummery := tapErrors.getErrorsSummary()

		logger.Log.Infof("%v (errors: %v, errTypes:%v) - Errors Summary: %s",
			time.Since(appStats.StartTime),
			tapErrors.nErrors,
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
}

func dumpMemoryProfile(filename string) error {
	if filename == "" {
		return nil
	}

	f, err := os.Create(*memprofile)

	if err != nil {
		return err
	}

	defer f.Close()

	if err := pprof.WriteHeapProfile(f); err != nil {
		return err
	}

	return nil
}

func printStatsSummary() {
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
	logger.Log.Infof("Errors: %d", tapErrors.nErrors)
	for e := range tapErrors.errorsMap {
		logger.Log.Infof(" %s:\t\t%d", e, tapErrors.errorsMap[e])
	}
	logger.Log.Infof("AppStats: %v", GetStats())
}

func startPassiveTapper(outputItems chan *api.OutputChannelItem) {
	streamsMap := NewTcpStreamMap()
	go streamsMap.closeTimedoutTcpStreamChannels()

	var outputLevel int

	defer util.Run()()
	if *debug {
		outputLevel = 2
	} else if *verbose {
		outputLevel = 1
	} else if *quiet {
		outputLevel = -1
	}

	tapErrors = NewErrorsMap(outputLevel)

	var bpffilter string
	if len(flag.Args()) > 0 {
		bpffilter = strings.Join(flag.Args(), " ")
	}

	packetSource, err := NewTcpPacketSource(0, *fname, *iface, tcpPacketSourceBehaviour{
		snapLength:  *snaplen,
		promisc:     *promisc,
		tstype:      *tstype,
		decoderName: *decoder,
		lazy:        *lazy,
		bpfFilter:   bpffilter,
	})

	if err != nil {
		logger.Log.Fatal(err)
	}

	defer packetSource.close()

	packets := make(chan tcpPacketInfo, 10000)
	assembler := NewTcpAssember(outputItems, streamsMap)

	logger.Log.Info("Starting to read packets")
	appStats.SetStartTime(time.Now())

	go packetSource.readPackets(!*nodefrag, packets)

	staleConnectionTimeout := time.Second * time.Duration(*staleTimeoutSeconds)
	cleaner := Cleaner{
		assembler:         assembler.Assembler,
		assemblerMutex:    &assembler.assemblerMutex,
		cleanPeriod:       cleanPeriod,
		connectionTimeout: staleConnectionTimeout,
	}
	cleaner.start()

	go printPeriodicStats(&cleaner)

	if GetMemoryProfilingEnabled() {
		startMemoryProfiler()
	}

	assembler.processPackets(*hexdumppkt, packets)

	if outputLevel >= 2 {
		assembler.dumpStreamPool()
	}

	if err := dumpMemoryProfile(*memprofile); err != nil {
		logger.Log.Errorf("Error dumping memory profile %v\n", err)
	}

	assembler.waitAndDump()

	printStatsSummary()
}
