package configStructs

import (
	"fmt"
	"github.com/up9inc/mizu/cli/mizu"
)

type ConfigStruct struct {
	Tap       TapConfig     `yaml:"tap"`
	Fetch     FetchConfig   `yaml:"fetch"`
	Version   VersionConfig `yaml:"version"`
	View      ViewConfig    `yaml:"view"`
	MizuImage string        `yaml:"mizu-image"`
	Telemetry bool          `yaml:"telemetry" default:"true"`
}

func (config *ConfigStruct) SetDefaults() {
	config.MizuImage = fmt.Sprintf("gcr.io/up9-docker-hub/mizu/%s:%s", mizu.Branch, mizu.SemVer)
}
