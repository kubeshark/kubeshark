package source

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/up9inc/mizu/shared/logger"
	"github.com/vishvananda/netns"
)

type PacketSourceManager struct {
	sources []*tcpPacketSource
}

func NewPacketSourceManager(procfs string, pids string, filename string, interfaceName string,
	behaviour TcpPacketSourceBehaviour) (*PacketSourceManager, error) {

	sources := make([]*tcpPacketSource, 0)
	host_source, err := newTcpPacketSource(filename, interfaceName, behaviour)

	if err != nil {
		return nil, err
	}

	sources = append(sources, host_source)

	if pids != "" {
		netnsSources := newNetnsPacketSources(procfs, pids, interfaceName, behaviour)

		if err != nil {
			sources = append(sources, netnsSources...)
		}
	}

	return &PacketSourceManager{
		sources: sources,
	}, nil
}

func newNetnsPacketSources(procfs string, pids string, interfaceName string,
	behaviour TcpPacketSourceBehaviour) []*tcpPacketSource {
	result := make([]*tcpPacketSource, 1)

	for _, pidstr := range strings.Split(pids, ",") {
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

	go func(done chan<- *tcpPacketSource) {
		if err := netns.Set(nsh); err != nil {
			logger.Log.Errorf("Unable to set netns of pid %v - %v", pid, err)
			close(done)
		}

		src, err := newTcpPacketSource("", interfaceName, behaviour)

		if err != nil {
			logger.Log.Errorf("Error listening to PID %v - %v", pid, err)
			close(done)
		}

		done <- src
	}(done)

	result, closed := <-done

	if closed {
		return nil, fmt.Errorf("unable to listen to PID: %v", pid)
	}

	return result, nil
}

func (m *PacketSourceManager) ReadPackets(ipdefrag bool, packets chan<- TcpPacketInfo) {
	for _, src := range m.sources {
		go src.readPackets(ipdefrag, packets)
	}
}

func (m *PacketSourceManager) Close() {
	for _, src := range m.sources {
		go src.close()
	}
}
