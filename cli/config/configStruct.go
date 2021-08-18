package config

import (
	"fmt"
	"github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/mizu"
	v1 "k8s.io/api/core/v1"
)

const (
	MizuResourcesNamespaceConfigName = "mizu-resources-namespace"
)

type ConfigStruct struct {
	Tap                    configStructs.TapConfig     `yaml:"tap"`
	Fetch                  configStructs.FetchConfig   `yaml:"fetch"`
	Version                configStructs.VersionConfig `yaml:"version"`
	View                   configStructs.ViewConfig    `yaml:"view"`
	Logs                   configStructs.LogsConfig    `yaml:"logs"`
	Config                 configStructs.ConfigConfig  `yaml:"config,omitempty"`
	AgentImage             string                      `yaml:"agent-image,omitempty" readonly:""`
	ImagePullPolicyStr     string                      `yaml:"image-pull-policy" default:"Always"`
	MizuResourcesNamespace string                      `yaml:"mizu-resources-namespace" default:"mizu"`
	Telemetry              bool                        `yaml:"telemetry" default:"true"`
	DumpLogs               bool                        `yaml:"dump-logs" default:"false"`
	KubeConfigPath         string                      `yaml:"kube-config-path"`
}

func (config *ConfigStruct) SetDefaults() {
	config.AgentImage = fmt.Sprintf("gcr.io/up9-docker-hub/mizu/%s:%s", mizu.Branch, mizu.SemVer)
}

func (config *ConfigStruct) ImagePullPolicy() v1.PullPolicy {
	return v1.PullPolicy(config.ImagePullPolicyStr)
}

func (config *ConfigStruct) IsNsRestrictedMode() bool {
	return config.MizuResourcesNamespace != "mizu" // Notice "mizu" string must match the default MizuResourcesNamespace
}
