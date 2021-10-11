package configStructs

import (
	"errors"
	"fmt"
	"github.com/up9inc/mizu/shared/units"
	"regexp"
)

const (
	GuiPortTapName                = "gui-port"
	NamespacesTapName             = "namespaces"
	AnalysisTapName               = "analysis"
	AllNamespacesTapName          = "all-namespaces"
	PlainTextFilterRegexesTapName = "regex-masking"
	DisableRedactionTapName       = "no-redact"
	HumanMaxEntriesDBSizeTapName  = "max-entries-db-size"
	DryRunTapName                 = "dry-run"
	WorkspaceTapName              = "workspace"
	EnforcePolicyFile             = "traffic-validation-file"
)

type TapConfig struct {
	UploadIntervalSec      int       `yaml:"upload-interval" default:"10"`
	PodRegexStr            string    `yaml:"regex" default:".*"`
	GuiPort                uint16    `yaml:"gui-port" default:"8899"`
	Namespaces             []string  `yaml:"namespaces"`
	Analysis               bool      `yaml:"analysis" default:"false"`
	AllNamespaces          bool      `yaml:"all-namespaces" default:"false"`
	PlainTextFilterRegexes []string  `yaml:"regex-masking"`
	IgnoredUserAgents      []string  `yaml:"ignored-user-agents"`
	DisableRedaction       bool      `yaml:"no-redact" default:"false"`
	HumanMaxEntriesDBSize  string    `yaml:"max-entries-db-size" default:"200MB"`
	DryRun                 bool      `yaml:"dry-run" default:"false"`
	Workspace              string    `yaml:"workspace"`
	EnforcePolicyFile      string    `yaml:"traffic-validation-file"`
	ApiServerResources     Resources `yaml:"api-server-resources"`
	TapperResources        Resources `yaml:"tapper-resources"`
}

type Resources struct {
	CpuLimit       string `yaml:"cpu-limit" default:"750m"`
	MemoryLimit    string `yaml:"memory-limit" default:"1Gi"`
	CpuRequests    string `yaml:"cpu-requests" default:"50m"`
	MemoryRequests string `yaml:"memory-requests" default:"50Mi"`
}

func (config *TapConfig) PodRegex() *regexp.Regexp {
	podRegex, _ := regexp.Compile(config.PodRegexStr)
	return podRegex
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

	if config.Workspace != "" {
		workspaceRegex, _ := regexp.Compile("[A-Za-z0-9][-A-Za-z0-9_.]*[A-Za-z0-9]+$")
		if len(config.Workspace) > 63 || !workspaceRegex.MatchString(config.Workspace) {
			return errors.New("invalid workspace name")
		}
	}

	if config.Analysis && config.Workspace != "" {
		return errors.New(fmt.Sprintf("Can't run with both --%s and --%s flags", AnalysisTapName, WorkspaceTapName))
	}

	return nil
}
