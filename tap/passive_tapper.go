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
	"strings"
	"time"
	"strconv"

	"github.com/shirou/gopsutil/cpu"
	"github.com/struCoder/pidusage"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/tap/api"
	"github.com/up9inc/mizu/tap/diagnose"
	"github.com/up9inc/mizu/tap/source"
	"github.com/up9inc/mizu/tap/tlstapper"
	v1 "k8s.io/api/core/v1"
)

const cleanPeriod = time.Second * 10

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
var ignoredPorts = flag.String("ignore-ports", "", "A comma separated list of ports to ignore")

// capture
var iface = flag.String("i", "en0", "Interface to read packets from")
var fname = flag.String("r", "", "Filename to read from, overrides -i")
var snaplen = flag.Int("s", 65536, "Snap length (number of bytes max to read per packet")
var tstype = flag.String("timestamp_type", "", "Type of timestamps to use")
var promisc = flag.Bool("promisc", true, "Set promiscuous mode")
var staleTimeoutSeconds = flag.Int("staletimout", 120, "Max time in seconds to keep connections which don't transmit data")
var servicemesh = flag.Bool("servicemesh", false, "Record decrypted traffic if the cluster is configured with a service mesh and with mtls")
var tls = flag.Bool("tls", false, "Enable TLS tapper")

var memprofile = flag.String("memprofile", "", "Write memory profile")

type TapOpts struct {
	HostMode         bool
	IgnoredPorts     []uint16
}

var extensions []*api.Extension                     // global
var filteringOptions *api.TrafficFilteringOptions   // global
var tapTargets []v1.Pod                             // global
var packetSourceManager *source.PacketSourceManager // global
var mainPacketInputChan chan source.TcpPacketInfo   // global
var tlsTapperInstance *tlstapper.TlsTapper          // global

func StartPassiveTapper(opts *TapOpts, outputItems chan *api.OutputChannelItem, extensionsRef []*api.Extension, options *api.TrafficFilteringOptions) {
	extensions = extensionsRef
	filteringOptions = options

	streamsMap := NewTcpStreamMap()

	if *tls {
		for _, e := range extensions {
			if e.Protocol.Name == "http" {
				tlsTapperInstance = startTlsTapper(e, outputItems, options, streamsMap)
				break
			}
		}
	}

	if GetMemoryProfilingEnabled() {
		diagnose.StartMemoryProfiler(os.Getenv(MemoryProfilingDumpPath), os.Getenv(MemoryProfilingTimeIntervalSeconds))
	}

	assembler := initializePassiveTapper(opts, outputItems, streamsMap)
	go startPassiveTapper(streamsMap, assembler)
}

func UpdateTapTargets(newTapTargets []v1.Pod) {
	success := true

	tapTargets = newTapTargets

	packetSourceManager.UpdatePods(tapTargets, !*nodefrag, mainPacketInputChan)

	if tlsTapperInstance != nil {
		if err := tlstapper.UpdateTapTargets(tlsTapperInstance, &tapTargets, *procfs); err != nil {
			tlstapper.LogError(err)
			success = false
		}
	}

	printNewTapTargets(success)
}

func printNewTapTargets(success bool) {
	printStr := ""
	for _, tapTarget := range tapTargets {
		printStr += fmt.Sprintf("%s (%s), ", tapTarget.Status.PodIP, tapTarget.Name)
	}
	printStr = strings.TrimRight(printStr, ", ")

	if success {
		logger.Log.Infof("Now tapping: %s", printStr)
	} else {
		logger.Log.Errorf("Failed to start tapping: %s", printStr)
	}
}

func printPeriodicStats(cleaner *Cleaner) {
	statsPeriod := time.Second * time.Duration(*statsevery)
	ticker := time.NewTicker(statsPeriod)

	logicalCoreCount, err := cpu.Counts(true)
	if err != nil {
		logicalCoreCount = -1
	}

	physicalCoreCount, err := cpu.Counts(false)
	if err != nil {
		physicalCoreCount = -1
	}

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
		sysInfo, err := pidusage.GetStat(os.Getpid())
		if err != nil {
			sysInfo = &pidusage.SysInfo{
				CPU:    -1,
				Memory: -1,
			}
		}
		logger.Log.Infof(
			"mem: %d, goroutines: %d, cpu: %f, cores: %d/%d, rss: %f",
			memStats.HeapAlloc,
			runtime.NumGoroutine(),
			sysInfo.CPU,
			logicalCoreCount,
			physicalCoreCount,
			sysInfo.Memory)

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
	packetSourceManager, err = source.NewPacketSourceManager(*procfs, *fname, *iface, *servicemesh, tapTargets, behaviour, !*nodefrag, mainPacketInputChan)
	return err
}

func initializePassiveTapper(opts *TapOpts, outputItems chan *api.OutputChannelItem, streamsMap api.TcpStreamMap) *tcpAssembler {
	diagnose.InitializeErrorsMap(*debug, *verbose, *quiet)
	diagnose.InitializeTapperInternalStats()

	mainPacketInputChan = make(chan source.TcpPacketInfo)

	if err := initializePacketSources(); err != nil {
		logger.Log.Fatal(err)
	}

	opts.IgnoredPorts = append(opts.IgnoredPorts, buildIgnoredPortsList(*ignoredPorts)...)

	assembler := NewTcpAssembler(outputItems, streamsMap, opts)

	return assembler
}

func startPassiveTapper(streamsMap api.TcpStreamMap, assembler *tcpAssembler) {
	go streamsMap.CloseTimedoutTcpStreamChannels()

	diagnose.AppStats.SetStartTime(time.Now())

	staleConnectionTimeout := time.Second * time.Duration(*staleTimeoutSeconds)
	cleaner := Cleaner{
		assembler:         assembler.Assembler,
		assemblerMutex:    &assembler.assemblerMutex,
		cleanPeriod:       cleanPeriod,
		connectionTimeout: staleConnectionTimeout,
		streamsMap:        streamsMap,
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

func startTlsTapper(extension *api.Extension, outputItems chan *api.OutputChannelItem,
	options *api.TrafficFilteringOptions, streamsMap api.TcpStreamMap) *tlstapper.TlsTapper {
	tls := tlstapper.TlsTapper{}
	chunksBufferSize := os.Getpagesize() * 100
	logBufferSize := os.Getpagesize()

	if err := tls.Init(chunksBufferSize, logBufferSize, *procfs, extension); err != nil {
		tlstapper.LogError(err)
		return nil
	}

	if err := tlstapper.UpdateTapTargets(&tls, &tapTargets, *procfs); err != nil {
		tlstapper.LogError(err)
		return nil
	}

	// A quick way to instrument libssl.so without PID filtering - used for debuging and troubleshooting
	//
	if os.Getenv("MIZU_GLOBAL_SSL_LIBRARY") != "" {
		if err := tls.GlobalTap(os.Getenv("MIZU_GLOBAL_SSL_LIBRARY")); err != nil {
			tlstapper.LogError(err)
			return nil
		}
	}

	var emitter api.Emitter = &api.Emitting{
		AppStats:      &diagnose.AppStats,
		OutputChannel: outputItems,
	}

	go tls.PollForLogging()
	go tls.Poll(emitter, options, streamsMap)

	return &tls
}

func buildIgnoredPortsList(ignoredPorts string) []uint16 {
	tmp := strings.Split(ignoredPorts, ",")
	result := make([]uint16, len(tmp))

	for i, raw := range tmp {
		v, err := strconv.Atoi(raw)
		if err != nil {
			continue
		}

		result[i] = uint16(v)
	}

	return result
}
