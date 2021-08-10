package config

import (
	"fmt"
	configStructs2 "github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/mizu"
)

const (
	AgentImageConfigName             = "agent-image"
	MizuResourcesNamespaceConfigName = "mizu-resources-namespace"
	TelemetryConfigName              = "telemetry"
	DumpLogsConfigName               = "dump-logs"
	KubeConfigPathName               = "kube-config-path"
)

type ConfigStruct struct {
	Tap                    configStructs2.TapConfig     `yaml:"tap"`
	Fetch                  configStructs2.FetchConfig   `yaml:"fetch"`
	Version                configStructs2.VersionConfig `yaml:"version"`
	View                   configStructs2.ViewConfig    `yaml:"view"`
	AgentImage             string                       `yaml:"agent-image,omitempty" readonly:""`
	MizuResourcesNamespace string                       `yaml:"mizu-resources-namespace" default:"mizu"`
	Telemetry              bool                         `yaml:"telemetry" default:"true"`
	DumpLogs               bool                         `yaml:"dump-logs" default:"false"`
	KubeConfigPath         string                       `yaml:"kube-config-path" default:""`
}

func (config *ConfigStruct) SetDefaults() {
	config.AgentImage = fmt.Sprintf("gcr.io/up9-docker-hub/mizu/%s:%s", mizu.Branch, mizu.SemVer)
}

func (config *ConfigStruct) IsNsRestrictedMode() bool {
	return config.MizuResourcesNamespace != "mizu" // Notice "mizu" string must match the default MizuResourcesNamespace
}
