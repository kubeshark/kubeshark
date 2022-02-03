package tlstapper

import (
	"bytes"
	"encoding/binary"

	"github.com/cilium/ebpf/perf"
	"github.com/cilium/ebpf/rlimit"
	"github.com/go-errors/errors"
	"github.com/up9inc/mizu/shared/logger"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go tlsTapper bpf/tls_tapper.c -- -O2 -g -D__TARGET_ARCH_x86

type TlsTapper struct {
	bpfObjects   tlsTapperObjects
	syscallHooks syscallHooks
	sslHooks     []sslHooks
	reader       *perf.Reader
}

func (t *TlsTapper) Init(bufferSize int) error {
	logger.Log.Infof("Initializing tls tapper (bufferSize: %v)", bufferSize)

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

	t.sslHooks = make([]sslHooks, 0)

	return t.initChunksReader(bufferSize)
}

func (t *TlsTapper) pollPerf(chunks chan<- *tlsChunk) {
	logger.Log.Infof("Start polling for tls events")

	for {
		record, err := t.reader.Read()

		if err != nil {
			close(chunks)

			if errors.Is(err, perf.ErrClosed) {
				return
			}

			LogError(errors.Errorf("Error reading chunks from tls perf, aborting TLS! %v", err))
			return
		}

		if record.LostSamples != 0 {
			logger.Log.Infof("Buffer is full, dropped %d chunks", record.LostSamples)
			continue
		}

		buffer := bytes.NewReader(record.RawSample)

		var chunk tlsChunk

		if err := binary.Read(buffer, binary.LittleEndian, &chunk); err != nil {
			LogError(errors.Errorf("Error parsing chunk %v", err))
			continue
		}

		chunks <- &chunk
	}
}

func (t *TlsTapper) GlobalTap(sslLibrary string) error {
	return t.tapPid(0, sslLibrary)
}

func (t *TlsTapper) AddPid(procfs string, pid uint32) error {
	sslLibrary, err := findSsllib(procfs, pid)

	if err != nil {
		return err
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

	for _, sslHooks := range t.sslHooks {
		errors = append(errors, sslHooks.close()...)
	}

	if err := t.reader.Close(); err != nil {
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

func (t *TlsTapper) initChunksReader(bufferSize int) error {
	var err error

	t.reader, err = perf.NewReader(t.bpfObjects.ChunksBuffer, bufferSize)

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

	t.sslHooks = append(t.sslHooks, newSsl)

	pids := t.bpfObjects.tlsTapperMaps.PidsMap

	if err := pids.Put(pid, uint32(1)); err != nil {
		return errors.Wrap(err, 0)
	}

	return nil
}

func LogError(err error) {
	switch err := err.(type) {
	case *errors.Error:
		logger.Log.Errorf("Error: %v", err.ErrorStack())
	default:
		logger.Log.Errorf("Error: %v", err)
	}
}
