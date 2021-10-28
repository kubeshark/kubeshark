package config

import (
	"encoding/json"
	"fmt"
	"github.com/up9inc/mizu/shared"
	"io/ioutil"
)

var Config *shared.MizuConfig

func LoadConfig() error {
	if Config != nil {
		return nil
	}
	filePath := fmt.Sprintf("%s%s", shared.ConfigDirPath, shared.ConfigFileName)
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, &Config)
	if err != nil {
		return err
	}
	return nil
}
