package mizu

import (
	"os"
	"path"
)

var (
	SemVer                                    = "0.0.1"
	Branch                                    = "develop"
	GitCommitHash                             = "" // this var is overridden using ldflags in makefile when building
	BuildTimestamp                            = "" // this var is overridden using ldflags in makefile when building
	RBACVersion                               = "v1"
	DaemonModePersistentVolumeSizeBufferBytes = int64(500 * 1000 * 1000) //500mb
)

func GetMizuFolderPath() string {
	home, homeDirErr := os.UserHomeDir()
	if homeDirErr != nil {
		return ""
	}
	return path.Join(home, ".mizu")
}
