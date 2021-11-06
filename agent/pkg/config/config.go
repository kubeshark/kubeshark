package config

import (
	"encoding/json"
	"fmt"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/tap/api"
	"io/ioutil"
	"os"
)

// these values are used when the config.json file is not present
const (
	defaultMaxDatabaseSizeBytes int64  = 200 * 1000 * 1000
	defaultRegexTarget          string = ".*"
	DefaultDatabasePath         string = "./entries"
)

var Config *shared.MizuAgentConfig

func LoadConfig() error {
	if Config != nil {
		return nil
	}
	filePath := fmt.Sprintf("%s%s", shared.ConfigDirPath, shared.ConfigFileName)

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return applyDefaultConfig()
		}
		return err
	}

	if err = json.Unmarshal(content, &Config); err != nil {
		return err
	}
	return nil
}

func applyDefaultConfig() error {
	defaultConfig, err := getDefaultConfig()
	if err != nil {
		return err
	}
	Config = defaultConfig
	return nil
}

func getDefaultConfig() (*shared.MizuAgentConfig, error) {
	regex, err := api.CompileRegexToSerializableRegexp(defaultRegexTarget)
	if err != nil {
		return nil, err
	}
	return &shared.MizuAgentConfig{
		TapTargetRegex:    *regex,
		MaxDBSizeBytes:    defaultMaxDatabaseSizeBytes,
		AgentDatabasePath: DefaultDatabasePath,
		DaemonMode:        false,
	}, nil
}
