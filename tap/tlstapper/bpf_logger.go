package tlstapper

import (
	"bytes"
	"encoding/binary"
	"strings"

	"github.com/cilium/ebpf/perf"
	"github.com/go-errors/errors"
	"github.com/up9inc/mizu/logger"
)

const logPrefix = "[bpf] "

// The same consts defined in log.h
//
const logLevelError = 0
const logLevelInfo = 1
const logLevelDebug = 2

type logMessage struct {
	Level       uint32
	MessageCode uint32
	Arg1        uint64
	Arg2        uint64
	Arg3        uint64
}

type bpfLogger struct {
	logReader *perf.Reader
}

func newBpfLogger() *bpfLogger {
	return &bpfLogger{
		logReader: nil,
	}
}

func (p *bpfLogger) init(bpfObjects *tlsTapperObjects, bufferSize int) error {
	var err error

	p.logReader, err = perf.NewReader(bpfObjects.LogBuffer, bufferSize)

	if err != nil {
		return errors.Wrap(err, 0)
	}

	return nil
}

func (p *bpfLogger) close() error {
	return p.logReader.Close()
}

func (p *bpfLogger) poll() {
	logger.Log.Infof("Start polling for bpf logs")

	for {
		record, err := p.logReader.Read()

		if err != nil {
			if errors.Is(err, perf.ErrClosed) {
				return
			}

			LogError(errors.Errorf("Error reading from bpf logger perf buffer, aboring logger! %w", err))
			return
		}

		if record.LostSamples != 0 {
			logger.Log.Infof("Log buffer is full, dropped %d logs", record.LostSamples)
			continue
		}

		buffer := bytes.NewReader(record.RawSample)

		var log logMessage

		if err := binary.Read(buffer, binary.LittleEndian, &log); err != nil {
			LogError(errors.Errorf("Error parsing log %v", err))
			continue
		}

		p.log(&log)
	}
}

func (p *bpfLogger) log(log *logMessage) {
	if int(log.MessageCode) >= len(bpfLogMessages) {
		logger.Log.Errorf("Unknown message code from bpf logger %d", log.MessageCode)
		return
	}

	format := bpfLogMessages[log.MessageCode]
	tokensCount := strings.Count(format, "%")

	if tokensCount == 0 {
		p.logLevel(log.Level, format)
	} else if tokensCount == 1 {
		p.logLevel(log.Level, format, log.Arg1)
	} else if tokensCount == 2 {
		p.logLevel(log.Level, format, log.Arg1, log.Arg2)
	} else if tokensCount == 3 {
		p.logLevel(log.Level, format, log.Arg1, log.Arg2, log.Arg3)
	}
}

func (p *bpfLogger) logLevel(level uint32, format string, args ...interface{}) {
	if level == logLevelError {
		logger.Log.Errorf(logPrefix+format, args...)
	} else if level == logLevelInfo {
		logger.Log.Infof(logPrefix+format, args...)
	} else if level == logLevelDebug {
		logger.Log.Debugf(logPrefix+format, args...)
	}
}
