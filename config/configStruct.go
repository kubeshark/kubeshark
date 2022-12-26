package config

import (
	"os"
	"path"
	"path/filepath"

	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/kubeshark"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/homedir"
)

const (
	ResourcesNamespaceConfigName = "resources-namespace"
	ConfigFilePathCommandName    = "config-path"
	KubeConfigPathConfigName     = "kube-config-path"
)

func CreateDefaultConfig() ConfigStruct {
	config := ConfigStruct{}

	config.Tap.Hub = configStructs.HubConfig{
		SrcPort: 8898,
		DstPort: 80,
	}

	config.Tap.Front = configStructs.FrontConfig{
		SrcPort: 8899,
		DstPort: 80,
	}

	return config
}

type ConfigStruct struct {
	Tap                configStructs.TapConfig    `yaml:"tap"`
	Logs               configStructs.LogsConfig   `yaml:"logs"`
	Config             configStructs.ConfigConfig `yaml:"config,omitempty"`
	ImagePullPolicyStr string                     `yaml:"image-pull-policy" default:"Always"`
	ResourcesNamespace string                     `yaml:"resources-namespace" default:"kubeshark"`
	DumpLogs           bool                       `yaml:"dump-logs" default:"false"`
	KubeConfigPathStr  string                     `yaml:"kube-config-path"`
	KubeContext        string                     `yaml:"kube-context"`
	ConfigFilePath     string                     `yaml:"config-path,omitempty" readonly:""`
	HeadlessMode       bool                       `yaml:"headless" default:"false"`
}

func (config *ConfigStruct) SetDefaults() {
	config.ConfigFilePath = path.Join(kubeshark.GetKubesharkFolderPath(), "config.yaml")
}

func (config *ConfigStruct) ImagePullPolicy() v1.PullPolicy {
	return v1.PullPolicy(config.ImagePullPolicyStr)
}

func (config *ConfigStruct) IsNsRestrictedMode() bool {
	return config.ResourcesNamespace != "kubeshark" // Notice "kubeshark" string must match the default KubesharkResourcesNamespace
}

func (config *ConfigStruct) KubeConfigPath() string {
	if config.KubeConfigPathStr != "" {
		return config.KubeConfigPathStr
	}

	envKubeConfigPath := os.Getenv("KUBECONFIG")
	if envKubeConfigPath != "" {
		return envKubeConfigPath
	}

	home := homedir.HomeDir()
	return filepath.Join(home, ".kube", "config")
}
