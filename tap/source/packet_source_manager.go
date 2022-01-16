package source

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/up9inc/mizu/shared/logger"
	"github.com/vishvananda/netns"
	v1 "k8s.io/api/core/v1"
)

type PacketSourceManager struct {
	sources []*tcpPacketSource
}

func NewPacketSourceManager(procfs string, pids string, filename string, interfaceName string,
	mtls bool, pods []v1.Pod, behaviour TcpPacketSourceBehaviour) (*PacketSourceManager, error) {
	sources := make([]*tcpPacketSource, 0)
	sources, err := createHostSource(sources, filename, interfaceName, behaviour)

	if err != nil {
		return nil, err
	}

	sources = createSourcesFromPids(sources, procfs, pids, interfaceName, behaviour)
	sources = createSourcesFromEnvoy(sources, mtls, procfs, pods, interfaceName, behaviour)
	sources = createSourcesFromLinkerd(sources, mtls, procfs, pods, interfaceName, behaviour)

	return &PacketSourceManager{
		sources: sources,
	}, nil
}

func createHostSource(sources []*tcpPacketSource, filename string, interfaceName string,
	behaviour TcpPacketSourceBehaviour) ([]*tcpPacketSource, error) {
	hostSource, err := newHostPacketSource(filename, interfaceName, behaviour)

	if err != nil {
		return sources, err
	}

	return append(sources, hostSource), nil
}

func createSourcesFromPids(sources []*tcpPacketSource, procfs string, pids string,
	interfaceName string, behaviour TcpPacketSourceBehaviour) []*tcpPacketSource {
	if pids == "" {
		return sources
	}

	netnsSources := newNetnsPacketSources(procfs, strings.Split(pids, ","), interfaceName, behaviour)
	sources = append(sources, netnsSources...)
	return sources
}

func createSourcesFromEnvoy(sources []*tcpPacketSource, mtls bool, procfs string, pods []v1.Pod,
	interfaceName string, behaviour TcpPacketSourceBehaviour) []*tcpPacketSource {
	if !mtls {
		return sources
	}

	envoyPids, err := discoverRelevantEnvoyPids(procfs, pods)

	if err != nil {
		logger.Log.Warningf("Unable to discover envoy pids - %v", err)
		return sources
	}

	netnsSources := newNetnsPacketSources(procfs, envoyPids, interfaceName, behaviour)
	sources = append(sources, netnsSources...)

	return sources
}

func createSourcesFromLinkerd(sources []*tcpPacketSource, mtls bool, procfs string, pods []v1.Pod,
	interfaceName string, behaviour TcpPacketSourceBehaviour) []*tcpPacketSource {
	if !mtls {
		return sources
	}

	linkerdPids, err := discoverRelevantLinkerdPids(procfs, pods)

	if err != nil {
		logger.Log.Warningf("Unable to discover linkerd pids - %v", err)
		return sources
	}

	netnsSources := newNetnsPacketSources(procfs, linkerdPids, interfaceName, behaviour)
	sources = append(sources, netnsSources...)

	return sources
}

func newHostPacketSource(filename string, interfaceName string,
	behaviour TcpPacketSourceBehaviour) (*tcpPacketSource, error) {
	var name string

	if filename == "" {
		name = fmt.Sprintf("host-%v", interfaceName)
	} else {
		name = fmt.Sprintf("file-%v", filename)
	}

	source, err := newTcpPacketSource(name, filename, interfaceName, behaviour)

	if err != nil {
		return nil, err
	}

	return source, nil
}

func newNetnsPacketSources(procfs string, pids []string, interfaceName string,
	behaviour TcpPacketSourceBehaviour) []*tcpPacketSource {
	result := make([]*tcpPacketSource, 0)

	for _, pidstr := range pids {
		pid, err := strconv.Atoi(pidstr)

		if err != nil {
			logger.Log.Errorf("Invalid PID: %v - %v", pid, err)
			continue
		}

		nsh, err := netns.GetFromPath(fmt.Sprintf("%v/%v/ns/net", procfs, pid))

		if err != nil {
			logger.Log.Errorf("Unable to get netns of pid %v - %v", pid, err)
			continue
		}

		src, err := newNetnsPacketSource(pid, nsh, interfaceName, behaviour)

		if err != nil {
			logger.Log.Errorf("Error starting netns packet source for %v - %v", pid, err)
			continue
		}

		result = append(result, src)
	}

	return result
}

func newNetnsPacketSource(pid int, nsh netns.NsHandle, interfaceName string,
	behaviour TcpPacketSourceBehaviour) (*tcpPacketSource, error) {

	done := make(chan *tcpPacketSource)
	errors := make(chan error)

	go func(done chan<- *tcpPacketSource) {
		// Setting a netns should be done from a dedicated OS thread.
		//
		// goroutines are not really OS threads, we try to mimic the issue by
		//	locking the OS thread to this goroutine
		//
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		oldnetns, err := netns.Get()

		if err != nil {
			logger.Log.Errorf("Unable to get netns of current thread %v", err)
			errors <- err
			return
		}

		if err := netns.Set(nsh); err != nil {
			logger.Log.Errorf("Unable to set netns of pid %v - %v", pid, err)
			errors <- err
			return
		}

		name := fmt.Sprintf("netns-%v-%v", pid, interfaceName)
		src, err := newTcpPacketSource(name, "", interfaceName, behaviour)

		if err != nil {
			logger.Log.Errorf("Error listening to PID %v - %v", pid, err)
			errors <- err
			return
		}

		if err := netns.Set(oldnetns); err != nil {
			logger.Log.Errorf("Unable to set back netns of current thread %v", err)
			errors <- err
			return
		}

		done <- src
	}(done)

	select {
	case err := <-errors:
		return nil, err
	case source := <-done:
		return source, nil
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
