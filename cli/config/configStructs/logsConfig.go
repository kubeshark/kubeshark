package configStructs

import (
	"fmt"
	"os"
	"path"
)

const (
	FileLogsName = "file"
)

type LogsConfig struct {
	FileStr string `yaml:"file"`
}

func (config *LogsConfig) Validate() error {
	if config.FileStr == "" {
		_, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get PWD, %v (try using `mizu logs -f <full path dest zip file>)`", err)
		}
	}

	return nil
}

func (config *LogsConfig) FilePath() string {
	if config.FileStr == "" {
		pwd, _ := os.Getwd()
		return path.Join(pwd, "mizu_logs.zip")
	}

	return config.FileStr
}
