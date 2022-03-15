package tlstapper

import (
	"github.com/cilium/ebpf/rlimit"
	"github.com/go-errors/errors"
	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/tap/api"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go tlsTapper bpf/tls_tapper.c -- -O2 -g -D__TARGET_ARCH_x86

type TlsTapper struct {
	bpfObjects      tlsTapperObjects
	syscallHooks    syscallHooks
	sslHooksStructs []sslHooks
	poller          *tlsPoller
	bpfLogger       *bpfLogger
}

func (t *TlsTapper) Init(chunksBufferSize int, logBufferSize int, procfs string, extension *api.Extension) error {
	logger.Log.Infof("Initializing tls tapper (chunksSize: %d) (logSize: %d)", chunksBufferSize, logBufferSize)

	if err := setupRLimit(); err != nil {
		return err
	}

	t.bpfObjects = tlsTapperObjects{}
	if err := loadTlsTapperObjects(&t.bpfObjects, nil); err != nil {
		return errors.Wrap(err, 0)
	}

	t.syscallHooks = syscallHooks{}
	if err := t.syscallHooks.installSyscallHooks(&t.bpfObjects); err != nil {
		return err
	}

	t.sslHooksStructs = make([]sslHooks, 0)

	t.bpfLogger = newBpfLogger()
	if err := t.bpfLogger.init(&t.bpfObjects, logBufferSize); err != nil {
		return err
	}

	t.poller = newTlsPoller(t, extension, procfs)
	return t.poller.init(&t.bpfObjects, chunksBufferSize)
}

func (t *TlsTapper) Poll(emitter api.Emitter, options *api.TrafficFilteringOptions) {
	t.poller.poll(emitter, options)
}

func (t *TlsTapper) PollForLogging() {
	t.bpfLogger.poll()
}

func (t *TlsTapper) GlobalTap(sslLibrary string) error {
	return t.tapPid(0, sslLibrary)
}

func (t *TlsTapper) AddPid(procfs string, pid uint32) error {
	sslLibrary, err := findSsllib(procfs, pid)

	if err != nil {
		logger.Log.Infof("PID skipped no libssl.so found (pid: %d) %v", pid, err)
		return nil // hide the error on purpose, its OK for a process to not use libssl.so
	}

	return t.tapPid(pid, sslLibrary)
}

func (t *TlsTapper) RemovePid(pid uint32) error {
	logger.Log.Infof("Removing PID (pid: %v)", pid)

	pids := t.bpfObjects.tlsTapperMaps.PidsMap

	if err := pids.Delete(pid); err != nil {
		return errors.Wrap(err, 0)
	}

	return nil
}

func (t *TlsTapper) Close() []error {
	errors := make([]error, 0)

	if err := t.bpfObjects.Close(); err != nil {
		errors = append(errors, err)
	}

	errors = append(errors, t.syscallHooks.close()...)

	for _, sslHooks := range t.sslHooksStructs {
		errors = append(errors, sslHooks.close()...)
	}

	if err := t.bpfLogger.close(); err != nil {
		errors = append(errors, err)
	}

	if err := t.poller.close(); err != nil {
		errors = append(errors, err)
	}

	return errors
}

func setupRLimit() error {
	err := rlimit.RemoveMemlock()

	if err != nil {
		return errors.Wrap(err, 0)
	}

	return nil
}

func (t *TlsTapper) tapPid(pid uint32, sslLibrary string) error {
	logger.Log.Infof("Tapping TLS (pid: %v) (sslLibrary: %v)", pid, sslLibrary)

	newSsl := sslHooks{}

	if err := newSsl.installUprobes(&t.bpfObjects, sslLibrary); err != nil {
		return err
	}

	t.sslHooksStructs = append(t.sslHooksStructs, newSsl)

	pids := t.bpfObjects.tlsTapperMaps.PidsMap

	if err := pids.Put(pid, uint32(1)); err != nil {
		return errors.Wrap(err, 0)
	}

	return nil
}

func LogError(err error) {
	var e *errors.Error
	if errors.As(err, &e) {
		logger.Log.Errorf("Error: %v", e.ErrorStack())
	} else {
		logger.Log.Errorf("Error: %v", err)
	}
}
