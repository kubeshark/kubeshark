package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/kubeshark/kubeshark/cli/config/configStructs"
	"github.com/kubeshark/kubeshark/cli/kubeshark"
	"github.com/kubeshark/kubeshark/shared"
	"github.com/op/go-logging"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/homedir"
)

const (
	KubesharkResourcesNamespaceConfigName = "kubeshark-resources-namespace"
	ConfigFilePathCommandName             = "config-path"
	KubeConfigPathConfigName              = "kube-config-path"
)

type ConfigStruct struct {
	Tap                         configStructs.TapConfig     `yaml:"tap"`
	Check                       configStructs.CheckConfig   `yaml:"check"`
	Install                     configStructs.InstallConfig `yaml:"install"`
	Version                     configStructs.VersionConfig `yaml:"version"`
	View                        configStructs.ViewConfig    `yaml:"view"`
	Logs                        configStructs.LogsConfig    `yaml:"logs"`
	Config                      configStructs.ConfigConfig  `yaml:"config,omitempty"`
	AgentImage                  string                      `yaml:"agent-image,omitempty" readonly:""`
	ImagePullPolicyStr          string                      `yaml:"image-pull-policy" default:"Always"`
	KubesharkResourcesNamespace string                      `yaml:"kubeshark-resources-namespace" default:"kubeshark"`
	DumpLogs                    bool                        `yaml:"dump-logs" default:"false"`
	KubeConfigPathStr           string                      `yaml:"kube-config-path"`
	KubeContext                 string                      `yaml:"kube-context"`
	ConfigFilePath              string                      `yaml:"config-path,omitempty" readonly:""`
	HeadlessMode                bool                        `yaml:"headless" default:"false"`
	LogLevelStr                 string                      `yaml:"log-level,omitempty" default:"INFO" readonly:""`
	ServiceMap                  bool                        `yaml:"service-map" default:"true"`
	OAS                         shared.OASConfig            `yaml:"oas"`
}

func (config *ConfigStruct) validate() error {
	if _, err := logging.LogLevel(config.LogLevelStr); err != nil {
		return fmt.Errorf("%s is not a valid log level, err: %v", config.LogLevelStr, err)
	}

	return nil
}

func (config *ConfigStruct) SetDefaults() {
	config.AgentImage = fmt.Sprintf("%s:%s", shared.KubesharkAgentImageRepo, kubeshark.Ver)
	config.ConfigFilePath = path.Join(kubeshark.GetKubesharkFolderPath(), "config.yaml")
}

func (config *ConfigStruct) ImagePullPolicy() v1.PullPolicy {
	return v1.PullPolicy(config.ImagePullPolicyStr)
}

func (config *ConfigStruct) IsNsRestrictedMode() bool {
	return config.KubesharkResourcesNamespace != "kubeshark" // Notice "kubeshark" string must match the default KubesharkResourcesNamespace
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

func (config *ConfigStruct) LogLevel() logging.Level {
	logLevel, _ := logging.LogLevel(config.LogLevelStr)
	return logLevel
}
