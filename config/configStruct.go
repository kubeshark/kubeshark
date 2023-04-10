package config

import (
	"os"
	"path/filepath"

	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/misc"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/homedir"
)

const (
	KubeConfigPathConfigName = "kube-configpath"
)

func CreateDefaultConfig() ConfigStruct {
	return ConfigStruct{}
}

type KubeConfig struct {
	ConfigPathStr string `yaml:"configpath"`
	Context       string `yaml:"context"`
}

type ManifestsConfig struct {
	Dump bool `yaml:"dump"`
}

type ConfigStruct struct {
	Tap               configStructs.TapConfig       `yaml:"tap"`
	Logs              configStructs.LogsConfig      `yaml:"logs"`
	Config            configStructs.ConfigConfig    `yaml:"config,omitempty"`
	Kube              KubeConfig                    `yaml:"kube"`
	DumpLogs          bool                          `yaml:"dumplogs" default:"false"`
	HeadlessMode      bool                          `yaml:"headless" default:"false"`
	License           string                        `yaml:"license" default:""`
	Scripting         configStructs.ScriptingConfig `yaml:"scripting"`
	ResourceLabels    map[string]string             `yaml:"resourceLabels" default:"{}"`
	NodeSelectorTerms []v1.NodeSelectorTerm         `yaml:"nodeSelectorTerms" default:"[]"`
	Manifests         ManifestsConfig               `yaml:"manifests,omitempty"`
}

func (config *ConfigStruct) ImagePullPolicy() v1.PullPolicy {
	return v1.PullPolicy(config.Tap.Docker.ImagePullPolicy)
}

func (config *ConfigStruct) ImagePullSecrets() []v1.LocalObjectReference {
	var ref []v1.LocalObjectReference
	for _, name := range config.Tap.Docker.ImagePullSecrets {
		ref = append(ref, v1.LocalObjectReference{Name: name})
	}

	return ref
}

func (config *ConfigStruct) IsNsRestrictedMode() bool {
	return config.Tap.SelfNamespace != misc.Program // Notice "kubeshark" string must match the default SelfNamespace
}

func (config *ConfigStruct) KubeConfigPath() string {
	if config.Kube.ConfigPathStr != "" {
		return config.Kube.ConfigPathStr
	}

	envKubeConfigPath := os.Getenv("KUBECONFIG")
	if envKubeConfigPath != "" {
		return envKubeConfigPath
	}

	home := homedir.HomeDir()
	return filepath.Join(home, ".kube", "config")
}
