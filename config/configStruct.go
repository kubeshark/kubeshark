package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/kubeshark"
	"github.com/kubeshark/worker/models"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/homedir"
)

const (
	ResourcesNamespaceConfigName = "resources-namespace"
	ConfigFilePathCommandName    = "config-path"
	KubeConfigPathConfigName     = "kube-config-path"
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
	Hub                HubConfig                  `yaml:"hub"`
	Front              FrontConfig                `yaml:"front"`
	Tap                configStructs.TapConfig    `yaml:"tap"`
	Check              configStructs.CheckConfig  `yaml:"check"`
	View               configStructs.ViewConfig   `yaml:"view"`
	Logs               configStructs.LogsConfig   `yaml:"logs"`
	Config             configStructs.ConfigConfig `yaml:"config,omitempty"`
	ImagePullPolicyStr string                     `yaml:"image-pull-policy" default:"Always"`
	ResourcesNamespace string                     `yaml:"resources-namespace" default:"kubeshark"`
	DumpLogs           bool                       `yaml:"dump-logs" default:"false"`
	KubeConfigPathStr  string                     `yaml:"kube-config-path"`
	KubeContext        string                     `yaml:"kube-context"`
	ConfigFilePath     string                     `yaml:"config-path,omitempty" readonly:""`
	HeadlessMode       bool                       `yaml:"headless" default:"false"`
	LogLevelStr        string                     `yaml:"log-level,omitempty" default:"info" readonly:""`
	ServiceMap         bool                       `yaml:"service-map" default:"true"`
	OAS                models.OASConfig           `yaml:"oas"`
}

func (config *ConfigStruct) validate() error {
	if _, err := zerolog.ParseLevel(config.LogLevelStr); err != nil {
		return fmt.Errorf("%s is not a valid log level, err: %v", config.LogLevelStr, err)
	}

	return nil
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

func (config *ConfigStruct) LogLevel() zerolog.Level {
	logLevel, err := zerolog.ParseLevel(config.LogLevelStr)
	if err != nil {
		log.Error().Err(err).Str("log-level", config.LogLevelStr).Msg("Invalid log level")
	}
	return logLevel
}
