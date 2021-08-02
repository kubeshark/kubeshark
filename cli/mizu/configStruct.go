package mizu

import (
	"fmt"
	"regexp"
)

type ConfigStruct struct {
	Tap       TapConfig     `yaml:"tap"`
	Fetch     FetchConfig   `yaml:"fetch"`
	Version   VersionConfig `yaml:"version"`
	View      ViewConfig    `yaml:"view"`
	MizuImage string        `yaml:"mizu-image"`
	Telemetry bool          `yaml:"telemetry" default:"true"`
}

func (config *ConfigStruct) SetDefaults() {
	config.MizuImage = fmt.Sprintf("gcr.io/up9-docker-hub/mizu/%s:%s", Branch, SemVer)
}

type TapConfig struct {
	AnalysisDestination    string   `yaml:"dest" default:"up9.app"`
	SleepIntervalSec       int      `yaml:"upload-interval" default:"10"`
	GuiPort                uint16   `yaml:"gui-port" default:"8899"`
	Namespace              string   `yaml:"namespace"`
	Analysis               bool     `yaml:"analysis" default:"false"`
	AllNamespaces          bool     `yaml:"all-namespaces" default:"false"`
	KubeConfigPath         string   `yaml:"kube-config"`
	PlainTextFilterRegexes []string `yaml:"regex-masking"`
	HideHealthChecks       bool     `yaml:"hide-healthchecks" default:"false"`
	DisableRedaction       bool     `yaml:"no-redact" default:"false"`
	HumanMaxEntriesDBSize  string   `yaml:"max-entries-db-size" default:"200MB"`
	MaxEntriesDBSizeBytes  int64
	Direction              string `yaml:"direction" default:"in"`
	TapOutgoing            bool
	PodRegexStr            string `yaml:"regex" default:".*"`
	PodRegex               *regexp.Regexp
}

type FetchConfig struct {
	Directory     string `yaml:"directory" default:"."`
	FromTimestamp int    `yaml:"from" default:"0"`
	ToTimestamp   int    `yaml:"to" default:"0"`
	MizuPort      uint16 `yaml:"port" default:"8899"`
}

type VersionConfig struct {
	DebugInfo bool `yaml:"debug" default:"false"`
}

type ViewConfig struct {
	GuiPort        uint16 `yaml:"gui-port" default:"8899"`
	KubeConfigPath string `yaml:"kube-config"`
}
