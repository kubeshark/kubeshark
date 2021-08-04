package mizu

import (
	"fmt"
	"github.com/up9inc/mizu/cli/mizu/configStructs"
)

type ConfigStruct struct {
	Tap           configStructs.TapConfig     `yaml:"tap"`
	Fetch         configStructs.FetchConfig   `yaml:"fetch"`
	Version       configStructs.VersionConfig `yaml:"version"`
	View          configStructs.ViewConfig    `yaml:"view"`
	MizuImage     string                      `yaml:"mizu-image"`
	MizuNamespace string                      `yaml:"mizu-namespace"`
	Telemetry     bool                        `yaml:"telemetry" default:"true"`
}

func (config *ConfigStruct) SetDefaults() {
	config.MizuImage = fmt.Sprintf("gcr.io/up9-docker-hub/mizu/%s:%s", Branch, SemVer)
}

func (config *ConfigStruct) ResourcesNamespace() string {
	if config.MizuNamespace == "" {
		return ResourcesDefaultNamespace
	}

	return config.MizuNamespace
}

func (config *ConfigStruct) IsOwnNamespace() bool {
	if config.MizuNamespace == "" {
		return true
	}

	return false
}
