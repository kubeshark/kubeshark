package logger

import (
	"github.com/op/go-logging"
	"github.com/up9inc/mizu/cli/mizu"
	"os"
	"path"
)

var Log = logging.MustGetLogger("mizu_cli")

var format = logging.MustStringFormatter(
	`%{time} %{level:.5s} ▶ %{pid} %{shortfile} %{shortfunc} ▶ %{message}`,
)

func GetLogFilePath() string {
	return path.Join(mizu.GetMizuFolderPath(), "mizu_cli.log")
}

func InitLogger() {
	logPath := GetLogFilePath()
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

	Log.Debugf("\n\n\n")
	Log.Debugf("Running mizu version %v", mizu.SemVer)
}
