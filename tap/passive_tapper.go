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
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/tap/api"
	"github.com/up9inc/mizu/tap/diagnose"
	"github.com/up9inc/mizu/tap/source"
	v1 "k8s.io/api/core/v1"
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
var procfs = flag.String("procfs", "/proc", "The procfs directory, used when mapping host volumes into a container")

// capture
var iface = flag.String("i", "en0", "Interface to read packets from")
var fname = flag.String("r", "", "Filename to read from, overrides -i")
var snaplen = flag.Int("s", 65536, "Snap length (number of bytes max to read per packet")
var tstype = flag.String("timestamp_type", "", "Type of timestamps to use")
var promisc = flag.Bool("promisc", true, "Set promiscuous mode")
var staleTimeoutSeconds = flag.Int("staletimout", 120, "Max time in seconds to keep connections which don't transmit data")
var pids = flag.String("pids", "", "A comma separated list of PIDs to capture their network namespaces")
var istio = flag.Bool("istio", false, "Record decrypted traffic if the cluster configured with istio and mtls")

var memprofile = flag.String("memprofile", "", "Write memory profile")

type TapOpts struct {
	HostMode          bool
	FilterAuthorities []v1.Pod
}

var extensions []*api.Extension                     // global
var filteringOptions *api.TrafficFilteringOptions   // global
var tapTargets []v1.Pod                             // global
var packetSourceManager *source.PacketSourceManager // global
var mainPacketInputChan chan source.TcpPacketInfo   // global

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

func StartPassiveTapper(opts *TapOpts, outputItems chan *api.OutputChannelItem, extensionsRef []*api.Extension, options *api.TrafficFilteringOptions) {
	extensions = extensionsRef
	filteringOptions = options

	if opts.FilterAuthorities == nil {
		tapTargets = []v1.Pod{}
	} else {
		tapTargets = opts.FilterAuthorities
	}

	if GetMemoryProfilingEnabled() {
		diagnose.StartMemoryProfiler(os.Getenv(MemoryProfilingDumpPath), os.Getenv(MemoryProfilingTimeIntervalSeconds))
	}

	go startPassiveTapper(opts, outputItems)
}

func UpdateTapTargets(newTapTargets []v1.Pod) {
	tapTargets = newTapTargets
	initializePacketSources()
	printNewTapTargets()
}

func printNewTapTargets() {
	printStr := ""
	for _, tapTarget := range tapTargets {
		printStr += tapTarget.Status.PodIP + " "
	}
	logger.Log.Infof("Now tapping: %s", printStr)
}

func printPeriodicStats(cleaner *Cleaner) {
	statsPeriod := time.Second * time.Duration(*statsevery)
	ticker := time.NewTicker(statsPeriod)

	for {
		<-ticker.C

		// Since the start
		errorMapLen, errorsSummery := diagnose.TapErrors.GetErrorsSummary()

		logger.Log.Infof("%v (errors: %v, errTypes:%v) - Errors Summary: %s",
			time.Since(diagnose.AppStats.StartTime),
			diagnose.TapErrors.ErrorsCount,
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
		currentAppStats := diagnose.AppStats.DumpStats()
		appStatsJSON, _ := json.Marshal(currentAppStats)
		logger.Log.Infof("app stats - %v", string(appStatsJSON))
	}
}

func initializePacketSources() error {
	if packetSourceManager != nil {
		packetSourceManager.Close()
	}

	var bpffilter string
	if len(flag.Args()) > 0 {
		bpffilter = strings.Join(flag.Args(), " ")
	}

	behaviour := source.TcpPacketSourceBehaviour{
		SnapLength:  *snaplen,
		Promisc:     *promisc,
		Tstype:      *tstype,
		DecoderName: *decoder,
		Lazy:        *lazy,
		BpfFilter:   bpffilter,
	}

	var err error
	if packetSourceManager, err = source.NewPacketSourceManager(*procfs, *pids, *fname, *iface, *istio, tapTargets, behaviour); err != nil {
		return err
	} else {
		packetSourceManager.ReadPackets(!*nodefrag, mainPacketInputChan)
		return nil
	}
}

func startPassiveTapper(opts *TapOpts, outputItems chan *api.OutputChannelItem) {
	streamsMap := NewTcpStreamMap()
	go streamsMap.closeTimedoutTcpStreamChannels()

	diagnose.InitializeErrorsMap(*debug, *verbose, *quiet)
	diagnose.InitializeTapperInternalStats()

	err := initializePacketSources()

	if err != nil {
		logger.Log.Fatal(err)
	}

	mainPacketInputChan = make(chan source.TcpPacketInfo)
	assembler := NewTcpAssembler(outputItems, streamsMap, opts)

	diagnose.AppStats.SetStartTime(time.Now())

	staleConnectionTimeout := time.Second * time.Duration(*staleTimeoutSeconds)
	cleaner := Cleaner{
		assembler:         assembler.Assembler,
		assemblerMutex:    &assembler.assemblerMutex,
		cleanPeriod:       cleanPeriod,
		connectionTimeout: staleConnectionTimeout,
	}
	cleaner.start()

	go printPeriodicStats(&cleaner)

	assembler.processPackets(*hexdumppkt, mainPacketInputChan)

	if diagnose.TapErrors.OutputLevel >= 2 {
		assembler.dumpStreamPool()
	}

	if err := diagnose.DumpMemoryProfile(*memprofile); err != nil {
		logger.Log.Errorf("Error dumping memory profile %v", err)
	}

	assembler.waitAndDump()

	diagnose.InternalStats.PrintStatsSummary()
	diagnose.TapErrors.PrintSummary()
	logger.Log.Infof("AppStats: %v", diagnose.AppStats)
}
