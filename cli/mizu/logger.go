package mizu

import (
	"fmt"
	"github.com/op/go-logging"
	"os"
	"path"
)

var Log = logging.MustGetLogger("mizu_cli")

var format = logging.MustStringFormatter(
	`%{time} %{level:.5s} ▶ %{pid} %{shortfile} %{shortfunc} ▶ %{message}`,
)

func InitLogger() {
	homeDirPath, _ := os.UserHomeDir()
	mizuDirPath := path.Join(homeDirPath, MizuFolderName)
	if err := os.MkdirAll(mizuDirPath, os.ModePerm); err != nil {
		panic(fmt.Sprintf("Failed creating mizu dir: %v, err %v", mizuDirPath, err))
	}
	logPath := path.Join(mizuDirPath, "log.log")
	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("Failed mizu log file: %v, err %v", logPath, err))
	}

	fileLog := logging.NewLogBackend(f, "", 0)
	consoleLog := logging.NewLogBackend(os.Stderr, "", 0)

	backend2Formatter := logging.NewBackendFormatter(fileLog, format)

	backend1Leveled := logging.AddModuleLevel(consoleLog)
	backend1Leveled.SetLevel(logging.INFO, "")

	logging.SetBackend(backend1Leveled, backend2Formatter)

	Log.Debugf("Running mizu version %v", SemVer)
}
