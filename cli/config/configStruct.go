package config

import (
	"fmt"
	"github.com/op/go-logging"
	"github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/mizu"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/homedir"
	"os"
	"path"
	"path/filepath"
)

const (
	MizuResourcesNamespaceConfigName = "mizu-resources-namespace"
	ConfigFilePathCommandName        = "config-path"
	KubeConfigPathConfigName         = "kube-config-path"
)

type ConfigStruct struct {
	Tap                    configStructs.TapConfig     `yaml:"tap"`
	Version                configStructs.VersionConfig `yaml:"version"`
	View                   configStructs.ViewConfig    `yaml:"view"`
	Logs                   configStructs.LogsConfig    `yaml:"logs"`
	Auth                   configStructs.AuthConfig    `yaml:"auth"`
	Config                 configStructs.ConfigConfig  `yaml:"config,omitempty"`
	AgentImage             string                      `yaml:"agent-image,omitempty" readonly:""`
	ImagePullPolicyStr     string                      `yaml:"image-pull-policy" default:"Always"`
	MizuResourcesNamespace string                      `yaml:"mizu-resources-namespace" default:"mizu"`
	Telemetry              bool                        `yaml:"telemetry" default:"true"`
	DumpLogs               bool                        `yaml:"dump-logs" default:"false"`
	KubeConfigPathStr      string                      `yaml:"kube-config-path"`
	ConfigFilePath         string                      `yaml:"config-path,omitempty" readonly:""`
	HeadlessMode           bool                        `yaml:"headless" default:"false"`
	LogLevelStr            string                      `yaml:"log-level,omitempty" default:"INFO" readonly:""`
	ServiceMap             bool                        `yaml:"service-map,omitempty" default:"false" readonly:""`
	OAS                    bool                        `yaml:"oas,omitempty" default:"false" readonly:""`
}

func (config *ConfigStruct) validate() error {
	if _, err := logging.LogLevel(config.LogLevelStr); err != nil {
		return fmt.Errorf("%s is not a valid log level, err: %v", config.LogLevelStr, err)
	}

	return nil
}

func (config *ConfigStruct) SetDefaults() {
	config.AgentImage = fmt.Sprintf("gcr.io/up9-docker-hub/mizu/%s:%s", mizu.Branch, mizu.SemVer)
	config.ConfigFilePath = path.Join(mizu.GetMizuFolderPath(), "config.yaml")
}

func (config *ConfigStruct) ImagePullPolicy() v1.PullPolicy {
	return v1.PullPolicy(config.ImagePullPolicyStr)
}

func (config *ConfigStruct) IsNsRestrictedMode() bool {
	return config.MizuResourcesNamespace != "mizu" // Notice "mizu" string must match the default MizuResourcesNamespace
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
