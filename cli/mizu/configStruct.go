package mizu

import (
	"fmt"

	"github.com/up9inc/mizu/cli/mizu/configStructs"
)

const (
	AgentImageConfigName             = "agent-image"
	MizuResourcesNamespaceConfigName = "mizu-resources-namespace"
	TelemetryConfigName              = "telemetry"
	DumpLogsConfigName               = "dump-logs"
)

type ConfigStruct struct {
	Tap                    configStructs.TapConfig     `yaml:"tap"`
	Fetch                  configStructs.FetchConfig   `yaml:"fetch"`
	Version                configStructs.VersionConfig `yaml:"version"`
	View                   configStructs.ViewConfig    `yaml:"view"`
	AgentImage             string                      `yaml:"agent-image"`
	MizuResourcesNamespace string                      `yaml:"mizu-resources-namespace" default:"mizu"`
	Telemetry              bool                        `yaml:"telemetry" default:"true"`
	DumpLogs               bool                        `yaml:"dump-logs" default:"false"`
}

func (config *ConfigStruct) SetDefaults() {
	config.AgentImage = fmt.Sprintf("gcr.io/up9-docker-hub/mizu/%s:%s", Branch, SemVer)
}

func (config *ConfigStruct) IsNsRestrictedMode() bool {
	return config.MizuResourcesNamespace != "mizu" // Notice "mizu" string must match the default MizuResourcesNamespace
}
