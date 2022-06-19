package tlstapper

import (
	"strconv"
	"sync"

	"github.com/cilium/ebpf/rlimit"
	"github.com/go-errors/errors"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/tap/api"
)

const GLOABL_TAP_PID = 0

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go@0d0727ef53e2f53b1731c73f4c61e0f58693083a -type tls_chunk tlsTapper bpf/tls_tapper.c -- -O2 -g -D__TARGET_ARCH_x86

type TlsTapper struct {
	bpfObjects      tlsTapperObjects
	syscallHooks    syscallHooks
	sslHooksStructs []sslHooks
	goHooksStructs  []goHooks
	poller          *tlsPoller
	bpfLogger       *bpfLogger
	registeredPids  sync.Map
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

	var err error
	t.poller, err = newTlsPoller(t, extension, procfs)

	if err != nil {
		return err
	}

	return t.poller.init(&t.bpfObjects, chunksBufferSize)
}

func (t *TlsTapper) Poll(emitter api.Emitter, options *api.TrafficFilteringOptions, streamsMap api.TcpStreamMap) {
	t.poller.poll(emitter, options, streamsMap)
}

func (t *TlsTapper) PollForLogging() {
	t.bpfLogger.poll()
}

func (t *TlsTapper) GlobalSsllibTap(sslLibrary string) error {
	return t.tapSsllibPid(GLOABL_TAP_PID, sslLibrary, api.UNKNOWN_NAMESPACE)
}

func (t *TlsTapper) GlobalGoTap(procfs string, pid string) error {
	_pid, err := strconv.Atoi(pid)
	if err != nil {
		return err
	}

	return t.tapGoPid(procfs, uint32(_pid), api.UNKNOWN_NAMESPACE)
}

func (t *TlsTapper) AddSsllibPid(procfs string, pid uint32, namespace string) error {
	sslLibrary, err := findSsllib(procfs, pid)

	if err != nil {
		logger.Log.Infof("PID skipped no libssl.so found (pid: %d) %v", pid, err)
		return nil // hide the error on purpose, its OK for a process to not use libssl.so
	}

	return t.tapSsllibPid(pid, sslLibrary, namespace)
}

func (t *TlsTapper) AddGoPid(procfs string, pid uint32, namespace string) error {
	return t.tapGoPid(procfs, pid, namespace)
}

func (t *TlsTapper) RemovePid(pid uint32) error {
	logger.Log.Infof("Removing PID (pid: %v)", pid)

	pids := t.bpfObjects.tlsTapperMaps.PidsMap

	if err := pids.Delete(pid); err != nil {
		return errors.Wrap(err, 0)
	}

	return nil
}

func (t *TlsTapper) ClearPids() {
	t.poller.clearPids()
	t.registeredPids.Range(func(key, v interface{}) bool {
		pid := key.(uint32)
		if pid == GLOABL_TAP_PID {
			return true
		}

		if err := t.RemovePid(pid); err != nil {
			LogError(err)
		}
		t.registeredPids.Delete(key)
		return true
	})
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

	for _, goHooks := range t.goHooksStructs {
		errors = append(errors, goHooks.close()...)
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

func (t *TlsTapper) tapSsllibPid(pid uint32, sslLibrary string, namespace string) error {
	logger.Log.Infof("Tapping TLS (pid: %v) (sslLibrary: %v)", pid, sslLibrary)

	newSsl := sslHooks{}

	if err := newSsl.installUprobes(&t.bpfObjects, sslLibrary); err != nil {
		return err
	}

	t.sslHooksStructs = append(t.sslHooksStructs, newSsl)

	t.poller.addPid(pid, namespace)

	pids := t.bpfObjects.tlsTapperMaps.PidsMap

	if err := pids.Put(pid, uint32(1)); err != nil {
		return errors.Wrap(err, 0)
	}

	t.registeredPids.Store(pid, true)

	return nil
}

func (t *TlsTapper) tapGoPid(procfs string, pid uint32, namespace string) error {
	exePath, err := findLibraryByPid(procfs, pid, "")
	if err != nil {
		return err
	}

	hooks := goHooks{}

	if err := hooks.installUprobes(&t.bpfObjects, exePath); err != nil {
		return err
	}

	logger.Log.Infof("Tapping TLS (pid: %v) (Go: %v)", pid, exePath)

	t.goHooksStructs = append(t.goHooksStructs, hooks)

	t.poller.addPid(pid, namespace)

	pids := t.bpfObjects.tlsTapperMaps.PidsMap

	if err := pids.Put(pid, uint32(1)); err != nil {
		return errors.Wrap(err, 0)
	}

	t.registeredPids.Store(pid, true)

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
