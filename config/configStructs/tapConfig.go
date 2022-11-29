package configStructs

import (
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"strings"

	"github.com/kubeshark/kubeshark/utils"
	"github.com/kubeshark/worker/models"
	"github.com/rs/zerolog/log"
)

const (
	ProxyPortLabel             = "proxy-port"
	NamespacesLabel            = "namespaces"
	AllNamespacesLabel         = "all-namespaces"
	EnableRedactionLabel       = "redact"
	HumanMaxEntriesDBSizeLabel = "max-entries-db-size"
	InsertionFilterName        = "insertion-filter"
	DryRunLabel                = "dry-run"
	ServiceMeshName            = "service-mesh"
	TlsName                    = "tls"
	ProfilerName               = "profiler"
	MaxLiveStreamsName         = "max-live-streams"
)

type DeployConfig struct {
	PodRegexStr       string   `yaml:"regex" default:".*"`
	ProxyPort         uint16   `yaml:"proxy-port" default:"8899"`
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
	HubResources          models.Resources `yaml:"hub-resources"`
	WorkerResources       models.Resources `yaml:"worker-resources"`
	ServiceMesh           bool             `yaml:"service-mesh" default:"false"`
	Tls                   bool             `yaml:"tls" default:"false"`
	PacketCapture         string           `yaml:"packet-capture" default:"libpcap"`
	Profiler              bool             `yaml:"profiler" default:"false"`
	MaxLiveStreams        int              `yaml:"max-live-streams" default:"500"`
}

func (config *DeployConfig) PodRegex() *regexp.Regexp {
	podRegex, _ := regexp.Compile(config.PodRegexStr)
	return podRegex
}

func (config *DeployConfig) MaxEntriesDBSizeBytes() int64 {
	maxEntriesDBSizeBytes, _ := utils.HumanReadableToBytes(config.HumanMaxEntriesDBSize)
	return maxEntriesDBSizeBytes
}

func (config *DeployConfig) GetInsertionFilter() string {
	insertionFilter := config.InsertionFilter
	if fs.ValidPath(insertionFilter) {
		if _, err := os.Stat(insertionFilter); err == nil {
			b, err := os.ReadFile(insertionFilter)
			if err != nil {
				log.Warn().Err(err).Str("insertion-filter-path", insertionFilter).Msg("Couldn't read the file! Defaulting to string.")
			} else {
				insertionFilter = string(b)
			}
		}
	}

	redactFilter := getRedactFilter(config)
	if insertionFilter != "" && redactFilter != "" {
		log.Info().Str("filter", insertionFilter).Msg("Using insertion filter:")
		return fmt.Sprintf("(%s) and (%s)", insertionFilter, redactFilter)
	} else if insertionFilter == "" && redactFilter != "" {
		return redactFilter
	}

	return insertionFilter
}

func getRedactFilter(config *DeployConfig) string {
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

func (config *DeployConfig) Validate() error {
	_, compileErr := regexp.Compile(config.PodRegexStr)
	if compileErr != nil {
		return fmt.Errorf("%s is not a valid regex %s", config.PodRegexStr, compileErr)
	}

	_, parseHumanDataSizeErr := utils.HumanReadableToBytes(config.HumanMaxEntriesDBSize)
	if parseHumanDataSizeErr != nil {
		return fmt.Errorf("Could not parse --%s value %s", HumanMaxEntriesDBSizeLabel, config.HumanMaxEntriesDBSize)
	}

	return nil
}
