package configStructs

import (
	"errors"
	"fmt"
	"github.com/up9inc/mizu/shared/units"
	"regexp"
	"strings"
)

const (
	GuiPortTapName                = "gui-port"
	NamespaceTapName              = "namespace"
	AnalysisTapName               = "analysis"
	AllNamespacesTapName          = "all-namespaces"
	KubeConfigPathTapName         = "kube-config"
	PlainTextFilterRegexesTapName = "regex-masking"
	HideHealthChecksTapName       = "hide-healthchecks"
	DisableRedactionTapName       = "no-redact"
	HumanMaxEntriesDBSizeTapName  = "max-entries-db-size"
	DirectionTapName              = "direction"
	TappedPodsPreviewTapName      = "pods-preview"
)

type TapConfig struct {
	AnalysisDestination    string   `yaml:"dest" default:"up9.app"`
	SleepIntervalSec       int      `yaml:"upload-interval" default:"10"`
	PodRegexStr            string   `yaml:"regex" default:".*"`
	GuiPort                uint16   `yaml:"gui-port" default:"8899"`
	Namespace              string   `yaml:"namespace"`
	Analysis               bool     `yaml:"analysis" default:"false"`
	AllNamespaces          bool     `yaml:"all-namespaces" default:"false"`
	KubeConfigPath         string   `yaml:"kube-config"`
	PlainTextFilterRegexes []string `yaml:"regex-masking"`
	HideHealthChecks       bool     `yaml:"hide-healthchecks" default:"false"`
	DisableRedaction       bool     `yaml:"no-redact" default:"false"`
	HumanMaxEntriesDBSize  string   `yaml:"max-entries-db-size" default:"200MB"`
	Direction              string   `yaml:"direction" default:"in"`
	TappedPodsPreview      bool     `yaml:"pods-preview" default:"false"`
}

func (config *TapConfig) PodRegex() *regexp.Regexp {
	podRegex, _ := regexp.Compile(config.PodRegexStr)
	return podRegex
}

func (config *TapConfig) TapOutgoing() bool {
	directionLowerCase := strings.ToLower(config.Direction)
	if directionLowerCase == "any" {
		return true
	}

	return false
}

func (config *TapConfig) MaxEntriesDBSizeBytes() int64 {
	maxEntriesDBSizeBytes, _ := units.HumanReadableToBytes(config.HumanMaxEntriesDBSize)
	return maxEntriesDBSizeBytes
}

func (config *TapConfig) Validate() error {
	_, compileErr := regexp.Compile(config.PodRegexStr)
	if compileErr != nil {
		return errors.New(fmt.Sprintf("%s is not a valid regex %s", config.PodRegexStr, compileErr))
	}

	_, parseHumanDataSizeErr := units.HumanReadableToBytes(config.HumanMaxEntriesDBSize)
	if parseHumanDataSizeErr != nil {
		return errors.New(fmt.Sprintf("Could not parse --%s value %s", HumanMaxEntriesDBSizeTapName, config.HumanMaxEntriesDBSize))
	}

	directionLowerCase := strings.ToLower(config.Direction)
	if directionLowerCase != "any" && directionLowerCase != "in" {
		return errors.New(fmt.Sprintf("%s is not a valid value for flag --%s. Acceptable values are in/any.", config.Direction, DirectionTapName))
	}

	return nil
}
