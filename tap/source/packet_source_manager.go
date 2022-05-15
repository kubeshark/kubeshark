package source

import (
	"fmt"
	"strings"

	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/tap/api"
	v1 "k8s.io/api/core/v1"
)

const bpfFilterMaxPods = 150
const hostSourcePid = "0"

type PacketSourceManagerConfig struct {
	mtls          bool
	procfs        string
	interfaceName string
	behaviour     TcpPacketSourceBehaviour
}

type PacketSourceManager struct {
	sources map[string]*tcpPacketSource
	config  PacketSourceManagerConfig
}

func NewPacketSourceManager(procfs string, filename string, interfaceName string,
	mtls bool, pods []v1.Pod, behaviour TcpPacketSourceBehaviour, ipdefrag bool, packets chan<- TcpPacketInfo) (*PacketSourceManager, error) {
	hostSource, err := newHostPacketSource(filename, interfaceName, behaviour)
	if err != nil {
		return nil, err
	}

	sourceManager := &PacketSourceManager{
		sources: map[string]*tcpPacketSource{
			hostSourcePid: hostSource,
		},
	}

	sourceManager.config = PacketSourceManagerConfig{
		mtls:          mtls,
		procfs:        procfs,
		interfaceName: interfaceName,
		behaviour:     behaviour,
	}

	go hostSource.readPackets(ipdefrag, packets)
	return sourceManager, nil
}

func newHostPacketSource(filename string, interfaceName string,
	behaviour TcpPacketSourceBehaviour) (*tcpPacketSource, error) {
	var name string
	if filename == "" {
		name = fmt.Sprintf("host-%s", interfaceName)
	} else {
		name = fmt.Sprintf("file-%s", filename)
	}

	source, err := newTcpPacketSource(name, filename, interfaceName, behaviour, api.Pcap)
	if err != nil {
		return nil, err
	}

	return source, nil
}

func (m *PacketSourceManager) UpdatePods(pods []v1.Pod, ipdefrag bool, packets chan<- TcpPacketInfo) {
	if m.config.mtls {
		m.updateMtlsPods(m.config.procfs, pods, m.config.interfaceName, m.config.behaviour, ipdefrag, packets)
	}

	m.setBPFFilter(pods)
}

func (m *PacketSourceManager) updateMtlsPods(procfs string, pods []v1.Pod,
	interfaceName string, behaviour TcpPacketSourceBehaviour, ipdefrag bool, packets chan<- TcpPacketInfo) {

	relevantPids := m.getRelevantPids(procfs, pods)
	logger.Log.Infof("Updating mtls pods (new: %v) (current: %v)", relevantPids, m.sources)

	for pid, src := range m.sources {
		if _, ok := relevantPids[pid]; !ok {
			src.close()
			delete(m.sources, pid)
		}
	}

	for pid, origin := range relevantPids {
		if _, ok := m.sources[pid]; !ok {
			source, err := newNetnsPacketSource(procfs, pid, interfaceName, behaviour, origin)

			if err == nil {
				go source.readPackets(ipdefrag, packets)
				m.sources[pid] = source
			}
		}
	}
}

func (m *PacketSourceManager) getRelevantPids(procfs string, pods []v1.Pod) map[string]api.Capture {
	relevantPids := make(map[string]api.Capture)
	relevantPids[hostSourcePid] = api.Pcap

	if envoyPids, err := discoverRelevantEnvoyPids(procfs, pods); err != nil {
		logger.Log.Warningf("Unable to discover envoy pids - %w", err)
	} else {
		for _, pid := range envoyPids {
			relevantPids[pid] = api.Envoy
		}
	}

	if linkerdPids, err := discoverRelevantLinkerdPids(procfs, pods); err != nil {
		logger.Log.Warningf("Unable to discover linkerd pids - %w", err)
	} else {
		for _, pid := range linkerdPids {
			relevantPids[pid] = api.Linkerd
		}
	}

	return relevantPids
}

func buildBPFExpr(pods []v1.Pod) string {
	hostsFilter := make([]string, 0)

	for _, pod := range pods {
		hostsFilter = append(hostsFilter, fmt.Sprintf("host %s", pod.Status.PodIP))
	}

	return fmt.Sprintf("%s and port not 443", strings.Join(hostsFilter, " or "))
}

func (m *PacketSourceManager) setBPFFilter(pods []v1.Pod) {
	if len(pods) == 0 {
		logger.Log.Info("No pods provided, skipping pcap bpf filter")
		return
	}

	var expr string

	if len(pods) > bpfFilterMaxPods {
		logger.Log.Info("Too many pods for setting ebpf filter %d, setting just not 443", len(pods))
		expr = "port not 443"
	} else {
		expr = buildBPFExpr(pods)
	}

	logger.Log.Infof("Setting pcap bpf filter %s", expr)

	for pid, src := range m.sources {
		if err := src.setBPFFilter(expr); err != nil {
			logger.Log.Warningf("Error setting bpf filter for %s %v - %w", pid, src, err)
		}
	}
}

func (m *PacketSourceManager) Close() {
	for _, src := range m.sources {
		src.close()
	}
}
