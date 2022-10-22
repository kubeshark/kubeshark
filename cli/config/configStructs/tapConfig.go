package configStructs

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/kubeshark/kubeshark/cli/uiUtils"
	"github.com/kubeshark/kubeshark/shared"

	"github.com/kubeshark/kubeshark/logger"
	"github.com/kubeshark/kubeshark/shared/units"
)

const (
	GuiPortTapName               = "gui-port"
	NamespacesTapName            = "namespaces"
	AllNamespacesTapName         = "all-namespaces"
	EnableRedactionTapName       = "redact"
	HumanMaxEntriesDBSizeTapName = "max-entries-db-size"
	InsertionFilterName          = "insertion-filter"
	DryRunTapName                = "dry-run"
	ServiceMeshName              = "service-mesh"
	TlsName                      = "tls"
	ProfilerName                 = "profiler"
	MaxLiveStreamsName           = "max-live-streams"
)

type TapConfig struct {
	PodRegexStr       string   `yaml:"regex" default:".*"`
	GuiPort           uint16   `yaml:"gui-port" default:"8899"`
	ProxyHost         string   `yaml:"proxy-host" default:"127.0.0.1"`
	Namespaces        []string `yaml:"namespaces"`
	AllNamespaces     bool     `yaml:"all-namespaces" default:"false"`
	IgnoredUserAgents []string `yaml:"ignored-user-agents"`
	EnableRedaction   bool     `yaml:"redact" default:"false"`
	RedactPatterns    struct {
		RequestHeaders     []string `yaml:"request-headers"`
		ResponseHeaders    []string `yaml:"response-headers"`
		RequestBody        []string `yaml:"request-body"`
		ResponseBody       []string `yaml:"response-body"`
		RequestQueryParams []string `yaml:"request-query-params"`
	} `yaml:"redact-patterns"`
	HumanMaxEntriesDBSize string           `yaml:"max-entries-db-size" default:"200MB"`
	InsertionFilter       string           `yaml:"insertion-filter" default:""`
	DryRun                bool             `yaml:"dry-run" default:"false"`
	ApiServerResources    shared.Resources `yaml:"api-server-resources"`
	TapperResources       shared.Resources `yaml:"tapper-resources"`
	ServiceMesh           bool             `yaml:"service-mesh" default:"false"`
	Tls                   bool             `yaml:"tls" default:"false"`
	PacketCapture         string           `yaml:"packet-capture" default:"libpcap"`
	Profiler              bool             `yaml:"profiler" default:"false"`
	MaxLiveStreams        int              `yaml:"max-live-streams" default:"500"`
}

func (config *TapConfig) PodRegex() *regexp.Regexp {
	podRegex, _ := regexp.Compile(config.PodRegexStr)
	return podRegex
}

func (config *TapConfig) MaxEntriesDBSizeBytes() int64 {
	maxEntriesDBSizeBytes, _ := units.HumanReadableToBytes(config.HumanMaxEntriesDBSize)
	return maxEntriesDBSizeBytes
}

func (config *TapConfig) GetInsertionFilter() string {
	insertionFilter := config.InsertionFilter
	if fs.ValidPath(insertionFilter) {
		if _, err := os.Stat(insertionFilter); err == nil {
			b, err := ioutil.ReadFile(insertionFilter)
			if err != nil {
				logger.Log.Warningf(uiUtils.Warning, fmt.Sprintf("Couldn't read the file on path: %s, err: %v", insertionFilter, err))
			} else {
				insertionFilter = string(b)
			}
		}
	}

	redactFilter := getRedactFilter(config)
	if insertionFilter != "" && redactFilter != "" {
		return fmt.Sprintf("(%s) and (%s)", insertionFilter, redactFilter)
	} else if insertionFilter == "" && redactFilter != "" {
		return redactFilter
	}

	return insertionFilter
}

func getRedactFilter(config *TapConfig) string {
	if !config.EnableRedaction {
		return ""
	}

	var redactValues []string
	for _, requestHeader := range config.RedactPatterns.RequestHeaders {
		redactValues = append(redactValues, fmt.Sprintf("request.headers['%s']", requestHeader))
	}
	for _, responseHeader := range config.RedactPatterns.ResponseHeaders {
		redactValues = append(redactValues, fmt.Sprintf("response.headers['%s']", responseHeader))
	}

	for _, requestBody := range config.RedactPatterns.RequestBody {
		redactValues = append(redactValues, fmt.Sprintf("request.postData.text.json()...%s", requestBody))
	}
	for _, responseBody := range config.RedactPatterns.ResponseBody {
		redactValues = append(redactValues, fmt.Sprintf("response.content.text.json()...%s", responseBody))
	}

	for _, requestQueryParams := range config.RedactPatterns.RequestQueryParams {
		redactValues = append(redactValues, fmt.Sprintf("request.queryString['%s']", requestQueryParams))
	}

	if len(redactValues) == 0 {
		return ""
	}

	return fmt.Sprintf("redact(\"%s\")", strings.Join(redactValues, "\",\""))
}

func (config *TapConfig) Validate() error {
	_, compileErr := regexp.Compile(config.PodRegexStr)
	if compileErr != nil {
		return fmt.Errorf("%s is not a valid regex %s", config.PodRegexStr, compileErr)
	}

	_, parseHumanDataSizeErr := units.HumanReadableToBytes(config.HumanMaxEntriesDBSize)
	if parseHumanDataSizeErr != nil {
		return fmt.Errorf("Could not parse --%s value %s", HumanMaxEntriesDBSizeTapName, config.HumanMaxEntriesDBSize)
	}

	return nil
}
