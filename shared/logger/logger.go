package logger

import (
	"os"

	"github.com/op/go-logging"
)

var Log = logging.MustGetLogger("mizu")

var format = logging.MustStringFormatter(
	`[%{time:2006-01-02T15:04:05.000-0700}] %{level:-5s} ▶ %{message} ▶ [%{pid} %{shortfile} %{shortfunc}]`,
)

func InitLogger(logPath string) {
	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		Log.Infof("Failed to open mizu log file: %v, err %v", logPath, err)
	}

	fileLog := logging.NewLogBackend(f, "", 0)
	consoleLog := logging.NewLogBackend(os.Stderr, "", 0)

	backend2Formatter := logging.NewBackendFormatter(fileLog, format)

	backend1Leveled := logging.AddModuleLevel(consoleLog)
	backend1Leveled.SetLevel(logging.INFO, "")

	logging.SetBackend(backend1Leveled, backend2Formatter)
}

func InitLoggerStd(level logging.Level) {
	var backends []logging.Backend

	stderrBackend := logging.NewLogBackend(os.Stderr, "", 0)
	stderrFormater := logging.NewBackendFormatter(stderrBackend, format)
	stderrLeveled := logging.AddModuleLevel(stderrFormater)
	stderrLeveled.SetLevel(logging.ERROR, "")
	backends = append(backends, stderrLeveled)

	if level >= logging.WARNING {
		stdoutBackend := logging.NewLogBackend(os.Stdout, "", 0)
		stdoutFormater := logging.NewBackendFormatter(stdoutBackend, format)
		stdoutLeveled := logging.AddModuleLevel(stdoutFormater)
		stdoutLeveled.SetLevel(level, "")
		backends = append(backends, stdoutLeveled)
	}
	logging.SetBackend(backends...)
}
