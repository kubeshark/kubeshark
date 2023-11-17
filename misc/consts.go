package misc

import (
	"fmt"
	"os"
	"path"
)

var (
	Software       = "Kubeshark"
	Program        = "kubeshark"
	Description    = "The API Traffic Analyzer for Kubernetes"
	Website        = "https://kubeshark.co"
	Email          = "info@kubeshark.co"
	Ver            = "0.0.0"
	Branch         = "master"
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
	return path.Join(home, fmt.Sprintf(".%s", Program))
}
