package logger

import (
	"os"

	"github.com/op/go-logging"
)

var moduleName = "mizu"
var Log = logging.MustGetLogger(moduleName)

var format = logging.MustStringFormatter(
	`%{time:2006-01-02T15:04:05.999Z-07:00} %{level:-5s} ▶ %{message} ▶ %{pid} %{shortfile} %{shortfunc}`,
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

func InitLoggerStderrOnly(level logging.Level) {
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)

	logging.SetBackend(backendFormatter)
	logging.SetLevel(level, "")
}

func SetLogLevel(level logging.Level) {
	logging.SetLevel(level, moduleName)
}
