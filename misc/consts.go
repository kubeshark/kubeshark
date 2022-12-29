package misc

import (
	"os"
	"path"
)

var (
	Ver            = "0.0"
	Branch         = "develop"
	GitCommitHash  = "" // this var is overridden using ldflags in makefile when building
	BuildTimestamp = "" // this var is overridden using ldflags in makefile when building
	RBACVersion    = "v1"
	Platform       = ""
)

func GetDotFolderPath() string {
	home, homeDirErr := os.UserHomeDir()
	if homeDirErr != nil {
		return ""
	}
	return path.Join(home, ".kubeshark")
}
