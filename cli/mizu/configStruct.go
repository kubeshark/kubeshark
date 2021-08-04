package mizu

import (
	"github.com/up9inc/mizu/cli/mizu/configStructs"
)

type ConfigStruct struct {
	Tap       configStructs.TapConfig     `yaml:"tap"`
	Fetch     configStructs.FetchConfig   `yaml:"fetch"`
	Version   configStructs.VersionConfig `yaml:"version"`
	View      configStructs.ViewConfig    `yaml:"view"`
	MizuImage string                      `yaml:"mizu-image"`
	Telemetry bool                        `yaml:"telemetry" default:"true"`
}

func (config *ConfigStruct) SetDefaults() {
	config.MizuImage = "gcr.io/sample-customer-264515/mizu-ui:latest"
}
