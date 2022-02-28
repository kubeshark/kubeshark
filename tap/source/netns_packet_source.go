package source

import (
	"fmt"
	"runtime"

	"github.com/up9inc/mizu/shared/logger"
	"github.com/vishvananda/netns"
)

func newNetnsPacketSource(procfs string, pid string,
	interfaceName string, behaviour TcpPacketSourceBehaviour) (*tcpPacketSource, error) {
	nsh, err := netns.GetFromPath(fmt.Sprintf("%v/%v/ns/net", procfs, pid))

	if err != nil {
		logger.Log.Errorf("Unable to get netns of pid %s - %w", pid, err)
		return nil, err
	}

	src, err := newPacketSourceFromNetnsHandle(pid, nsh, interfaceName, behaviour)

	if err != nil {
		logger.Log.Errorf("Error starting netns packet source for %s - %w", pid, err)
		return nil, err
	}

	return src, nil
}

func newPacketSourceFromNetnsHandle(pid string, nsh netns.NsHandle, interfaceName string,
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
			logger.Log.Errorf("Unable to get netns of current thread %w", err)
			errors <- err
			return
		}

		if err := netns.Set(nsh); err != nil {
			logger.Log.Errorf("Unable to set netns of pid %s - %w", pid, err)
			errors <- err
			return
		}

		name := fmt.Sprintf("netns-%v-%v", pid, interfaceName)
		src, err := newTcpPacketSource(name, "", interfaceName, behaviour)

		if err != nil {
			logger.Log.Errorf("Error listening to PID %s - %w", pid, err)
			errors <- err
			return
		}

		if err := netns.Set(oldnetns); err != nil {
			logger.Log.Errorf("Unable to set back netns of current thread %w", err)
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
