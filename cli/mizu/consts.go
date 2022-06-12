package mizu

import (
	"os"
	"path"
)

var (
	Ver                                        = "0.0"
	Branch                                     = "develop"
	GitCommitHash                              = "" // this var is overridden using ldflags in makefile when building
	BuildTimestamp                             = "" // this var is overridden using ldflags in makefile when building
	RBACVersion                                = "v1"
	Platform                                   = ""
)

const DEVENVVAR = "MIZU_DISABLE_TELEMTRY"

func GetMizuFolderPath() string {
	home, homeDirErr := os.UserHomeDir()
	if homeDirErr != nil {
		return ""
	}
	return path.Join(home, ".mizu")
}
