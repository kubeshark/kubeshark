package tlstapper

import (
	"github.com/cilium/ebpf/link"
	"github.com/go-errors/errors"
)

type tcpKprobeHooks struct {
	tcpSendmsg         link.Link
	tcpRecvmsg         link.Link
}

func (s *tcpKprobeHooks) installTcpKprobeHooks(bpfObjects *tlsTapperObjects) error {
	var err error

	s.tcpSendmsg, err = link.Kprobe("tcp_sendmsg", bpfObjects.TcpSendmsg, nil)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	s.tcpRecvmsg, err = link.Kprobe("tcp_recvmsg", bpfObjects.TcpRecvmsg, nil)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	return nil
}

func (s *tcpKprobeHooks) close() []error {
	returnValue := make([]error, 0)

	if s.tcpSendmsg != nil {
		if err := s.tcpSendmsg.Close(); err != nil {
			returnValue = append(returnValue, err)
		}
	}

	if s.tcpRecvmsg != nil {
		if err := s.tcpRecvmsg.Close(); err != nil {
			returnValue = append(returnValue, err)
		}
	}

	return returnValue
}
