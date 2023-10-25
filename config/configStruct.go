package config

import (
	"os"
	"path/filepath"

	"github.com/kubeshark/kubeshark/config/configStructs"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/homedir"
)

const (
	KubeConfigPathConfigName = "kube-configPath"
)

func CreateDefaultConfig() ConfigStruct {
	return ConfigStruct{
		Tap: configStructs.TapConfig{
			NodeSelectorTerms: []v1.NodeSelectorTerm{
				{
					MatchExpressions: []v1.NodeSelectorRequirement{
						{
							Key:      "kubernetes.io/os",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"linux"},
						},
					},
				},
			},
		},
	}
}

type KubeConfig struct {
	ConfigPathStr string `yaml:"configPath" json:"configPath"`
	Context       string `yaml:"context" json:"context"`
}

type ManifestsConfig struct {
	Dump bool `yaml:"dump" json:"dump"`
}

type ConfigStruct struct {
	Tap          configStructs.TapConfig       `yaml:"tap" json:"tap"`
	Logs         configStructs.LogsConfig      `yaml:"logs" json:"logs"`
	Config       configStructs.ConfigConfig    `yaml:"config,omitempty" json:"config,omitempty"`
	Kube         KubeConfig                    `yaml:"kube" json:"kube"`
	DumpLogs     bool                          `yaml:"dumpLogs" json:"dumpLogs" default:"false"`
	HeadlessMode bool                          `yaml:"headless" json:"headless" default:"false"`
	License      string                        `yaml:"license" json:"license" default:""`
	Scripting    configStructs.ScriptingConfig `yaml:"scripting" json:"scripting"`
	Manifests    ManifestsConfig               `yaml:"manifests,omitempty" json:"manifests,omitempty"`
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
