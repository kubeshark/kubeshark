package source

import (
	"fmt"
	"strings"

	"github.com/up9inc/mizu/shared/logger"
	v1 "k8s.io/api/core/v1"
)

const bpfFilterMaxPods = 150
const hostSourcePid = "0"

type PacketSourceManager struct {
	sources map[string]*tcpPacketSource
}

func NewPacketSourceManager(procfs string, filename string, interfaceName string,
	mtls bool, pods []v1.Pod, behaviour TcpPacketSourceBehaviour) (*PacketSourceManager, error) {
	hostSource, err := newHostPacketSource(filename, interfaceName, behaviour)
	if err != nil {
		return nil, err
	}

	sourceManager := &PacketSourceManager{
		sources: map[string]*tcpPacketSource{
			hostSourcePid: hostSource,
		},
	}

	sourceManager.UpdatePods(mtls, procfs, pods, interfaceName, behaviour)
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

	source, err := newTcpPacketSource(name, filename, interfaceName, behaviour)
	if err != nil {
		return nil, err
	}

	return source, nil
}

func (m *PacketSourceManager) UpdatePods(mtls bool, procfs string, pods []v1.Pod,
	interfaceName string, behaviour TcpPacketSourceBehaviour) {
	if mtls {
		m.updateMtlsPods(procfs, pods, interfaceName, behaviour)
	}

	m.setBPFFilter(pods)
}

func (m *PacketSourceManager) updateMtlsPods(procfs string, pods []v1.Pod,
	interfaceName string, behaviour TcpPacketSourceBehaviour) {

	relevantPids := m.getRelevantPids(procfs, pods)
	logger.Log.Infof("Updating mtls pods (new: %v) (current: %v)", relevantPids, m.sources)

	for pid, src := range m.sources {
		if _, ok := relevantPids[pid]; !ok {
			src.close()
			delete(m.sources, pid)
		}
	}

	for pid := range relevantPids {
		if _, ok := m.sources[pid]; !ok {
			source, err := newNetnsPacketSource(procfs, pid, interfaceName, behaviour)

			if err == nil {
				m.sources[pid] = source
			}
		}
	}
}

func (m *PacketSourceManager) getRelevantPids(procfs string, pods []v1.Pod) map[string]bool {
	relevantPids := make(map[string]bool)
	relevantPids[hostSourcePid] = true

	if envoyPids, err := discoverRelevantEnvoyPids(procfs, pods); err != nil {
		logger.Log.Warningf("Unable to discover envoy pids - %w", err)
	} else {
		for _, pid := range envoyPids {
			relevantPids[pid] = true
		}
	}

	if linkerdPids, err := discoverRelevantLinkerdPids(procfs, pods); err != nil {
		logger.Log.Warningf("Unable to discover linkerd pids - %w", err)
	} else {
		for _, pid := range linkerdPids {
			relevantPids[pid] = true
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

func (m *PacketSourceManager) ReadPackets(ipdefrag bool, packets chan<- TcpPacketInfo) {
	for _, src := range m.sources {
		go src.readPackets(ipdefrag, packets)
	}
}

func (m *PacketSourceManager) Close() {
	for _, src := range m.sources {
		src.close()
	}
}
