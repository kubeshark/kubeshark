package tlstapper

import (
	"strconv"
	"sync"

	"github.com/cilium/ebpf/rlimit"
	"github.com/go-errors/errors"
	"github.com/moby/moby/pkg/parsers/kernel"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/tap/api"
)

const GlobalTapPid = 0

// TODO: cilium/ebpf does not support .kconfig Therefore; for now, we build object files per kernel version.

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go@v0.9.0 -target $BPF_TARGET -cflags $BPF_CFLAGS -type tls_chunk -type goid_offsets tlsTapper bpf/tls_tapper.c

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go@v0.9.0 -target $BPF_TARGET -cflags "${BPF_CFLAGS} -DKERNEL_BEFORE_4_6" -type tls_chunk -type goid_offsets tlsTapper46 bpf/tls_tapper.c

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

	var err error
	err = SetupRLimit()
	if err != nil {
		return err
	}

	var kernelVersion *kernel.VersionInfo
	kernelVersion, err = kernel.GetKernelVersion()
	if err != nil {
		return err
	}

	logger.Log.Infof("Detected Linux kernel version: %s", kernelVersion)

	t.bpfObjects = tlsTapperObjects{}
	// TODO: cilium/ebpf does not support .kconfig Therefore; for now, we load object files according to kernel version.
	if kernel.CompareKernelVersion(*kernelVersion, kernel.VersionInfo{Kernel: 4, Major: 6, Minor: 0}) < 1 {
		if err := loadTlsTapper46Objects(&t.bpfObjects, nil); err != nil {
			return errors.Wrap(err, 0)
		}
	} else {
		if err := loadTlsTapperObjects(&t.bpfObjects, nil); err != nil {
			return errors.Wrap(err, 0)
		}
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

func (t *TlsTapper) GlobalSSLLibTap(sslLibrary string) error {
	return t.tapSSLLibPid(GlobalTapPid, sslLibrary, api.UnknownNamespace)
}

func (t *TlsTapper) GlobalGoTap(procfs string, pid string) error {
	_pid, err := strconv.Atoi(pid)
	if err != nil {
		return err
	}

	return t.tapGoPid(procfs, uint32(_pid), api.UnknownNamespace)
}

func (t *TlsTapper) AddSSLLibPid(procfs string, pid uint32, namespace string) error {
	sslLibrary, err := findSsllib(procfs, pid)

	if err != nil {
		logger.Log.Infof("PID skipped no libssl.so found (pid: %d) %v", pid, err)
		return nil // hide the error on purpose, it's OK for a process to not use libssl.so
	}

	return t.tapSSLLibPid(pid, sslLibrary, namespace)
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
		if pid == GlobalTapPid {
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
	returnValue := make([]error, 0)

	if err := t.bpfObjects.Close(); err != nil {
		returnValue = append(returnValue, err)
	}

	returnValue = append(returnValue, t.syscallHooks.close()...)

	for _, sslHooks := range t.sslHooksStructs {
		returnValue = append(returnValue, sslHooks.close()...)
	}

	for _, goHooks := range t.goHooksStructs {
		returnValue = append(returnValue, goHooks.close()...)
	}

	if err := t.bpfLogger.close(); err != nil {
		returnValue = append(returnValue, err)
	}

	if err := t.poller.close(); err != nil {
		returnValue = append(returnValue, err)
	}

	return returnValue
}

func SetupRLimit() error {
	err := rlimit.RemoveMemlock()

	if err != nil {
		return errors.Wrap(err, 0)
	}

	return nil
}

func (t *TlsTapper) tapSSLLibPid(pid uint32, sslLibrary string, namespace string) error {
	newSsl := sslHooks{}

	if err := newSsl.installUprobes(&t.bpfObjects, sslLibrary); err != nil {
		return err
	}

	logger.Log.Infof("Tapping TLS (pid: %v) (sslLibrary: %v)", pid, sslLibrary)

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
		logger.Log.Infof("PID skipped not a Go binary or symbol table is stripped (pid: %v) %v", pid, exePath)
		return nil // hide the error on purpose, its OK for a process to be not a Go binary or stripped Go binary
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
