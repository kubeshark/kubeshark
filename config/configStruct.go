package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/kubeshark"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/kubeshark/worker/models"
	"github.com/op/go-logging"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/homedir"
)

const (
	KubesharkResourcesNamespaceConfigName = "kubeshark-resources-namespace"
	ConfigFilePathCommandName             = "config-path"
	KubeConfigPathConfigName              = "kube-config-path"
)

type PortForward struct {
	SrcPort uint16 `yaml:"src-port"`
	DstPort uint16 `yaml:"dst-port"`
}

type HubConfig struct {
	PortForward PortForward `yaml:"port-forward"`
}

type FrontConfig struct {
	PortForward PortForward `yaml:"port-forward"`
}

func CreateDefaultConfig() ConfigStruct {
	config := ConfigStruct{}

	config.Hub = HubConfig{
		PortForward{
			8898,
			80,
		},
	}

	config.Front = FrontConfig{
		PortForward{
			8899,
			80,
		},
	}

	return config
}

type ConfigStruct struct {
	Hub                         HubConfig                   `yaml:"hub"`
	Front                       FrontConfig                 `yaml:"front"`
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
	OAS                         models.OASConfig            `yaml:"oas"`
}

func (config *ConfigStruct) validate() error {
	if _, err := logging.LogLevel(config.LogLevelStr); err != nil {
		return fmt.Errorf("%s is not a valid log level, err: %v", config.LogLevelStr, err)
	}

	return nil
}

func (config *ConfigStruct) SetDefaults() {
	config.AgentImage = fmt.Sprintf("%s:%s", utils.KubesharkAgentImageRepo, kubeshark.Ver)
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
