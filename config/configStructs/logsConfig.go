package configStructs

import (
	"fmt"
	"os"
	"path"

	"github.com/kubeshark/kubeshark/misc"
)

const (
	FileLogsName = "file"
	GrepLogsName = "grep"
)

type LogsConfig struct {
	FileStr string `yaml:"file" json:"file"`
	Grep    string `yaml:"grep" json:"grep"`
}

func (config *LogsConfig) Validate() error {
	if config.FileStr == "" {
		_, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get PWD, %v (try using `%s logs -f <full path dest zip file>)`", err, misc.Program)
		}
	}

	return nil
}

func (config *LogsConfig) FilePath() string {
	if config.FileStr == "" {
		pwd, _ := os.Getwd()
		return path.Join(pwd, fmt.Sprintf("%s_logs.zip", misc.Program))
	}

	return config.FileStr
}
