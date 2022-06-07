package tlstapper

import (
	"github.com/cilium/ebpf/link"
	"github.com/go-errors/errors"
)

type syscallHooks struct {
	sysEnterRead    link.Link
	sysEnterWrite   link.Link
	sysEnterAccept4 link.Link
	sysExitAccept4  link.Link
	sysEnterConnect link.Link
	sysExitConnect  link.Link
}

func (s *syscallHooks) installSyscallHooks(bpfObjects *tlsTapperObjects) error {
	var err error

	s.sysEnterRead, err = link.Tracepoint("syscalls", "sys_enter_read", bpfObjects.SysEnterRead)

	if err != nil {
		return errors.Wrap(err, 0)
	}

	s.sysEnterWrite, err = link.Tracepoint("syscalls", "sys_enter_write", bpfObjects.SysEnterWrite)

	if err != nil {
		return errors.Wrap(err, 0)
	}

	s.sysEnterAccept4, err = link.Tracepoint("syscalls", "sys_enter_accept4", bpfObjects.SysEnterAccept4)

	if err != nil {
		return errors.Wrap(err, 0)
	}

	s.sysExitAccept4, err = link.Tracepoint("syscalls", "sys_exit_accept4", bpfObjects.SysExitAccept4)

	if err != nil {
		return errors.Wrap(err, 0)
	}

	s.sysEnterConnect, err = link.Tracepoint("syscalls", "sys_enter_connect", bpfObjects.SysEnterConnect)

	if err != nil {
		return errors.Wrap(err, 0)
	}

	s.sysExitConnect, err = link.Tracepoint("syscalls", "sys_exit_connect", bpfObjects.SysExitConnect)

	if err != nil {
		return errors.Wrap(err, 0)
	}

	return nil
}

func (s *syscallHooks) close() []error {
	errors := make([]error, 0)

	if err := s.sysEnterRead.Close(); err != nil {
		errors = append(errors, err)
	}

	if err := s.sysEnterWrite.Close(); err != nil {
		errors = append(errors, err)
	}

	if err := s.sysEnterAccept4.Close(); err != nil {
		errors = append(errors, err)
	}

	if err := s.sysExitAccept4.Close(); err != nil {
		errors = append(errors, err)
	}

	if err := s.sysEnterConnect.Close(); err != nil {
		errors = append(errors, err)
	}

	if err := s.sysExitConnect.Close(); err != nil {
		errors = append(errors, err)
	}

	return errors
}
