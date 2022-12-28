package configStructs

import (
	"fmt"
	"regexp"

	"github.com/kubeshark/base/pkg/models"
	"github.com/kubeshark/kubeshark/utils"
)

const (
	DockerRegistryLabel        = "docker-registry"
	DockerTagLabel             = "docker-tag"
	ProxyPortFrontLabel        = "proxy-port-front"
	ProxyPortHubLabel          = "proxy-port-hub"
	ProxyHostLabel             = "proxy-host"
	NamespacesLabel            = "namespaces"
	AllNamespacesLabel         = "all-namespaces"
	HumanMaxEntriesDBSizeLabel = "max-entries-db-size"
	DryRunLabel                = "dry-run"
	PcapLabel                  = "pcap"
	ServiceMeshLabel           = "service-mesh"
	TlsLabel                   = "tls"
	DebugLabel                 = "debug"
)

type WorkerConfig struct {
	SrcPort uint16 `yaml:"src-port" default:"8897"`
	DstPort uint16 `yaml:"dst-port" default:"8897"`
}

type HubConfig struct {
	SrcPort uint16 `yaml:"src-port" default:"8898"`
	DstPort uint16 `yaml:"dst-port" default:"8898"`
}

type FrontConfig struct {
	SrcPort uint16 `yaml:"src-port" default:"8899"`
	DstPort uint16 `yaml:"dst-port" default:"80"`
}

type ProxyConfig struct {
	Worker WorkerConfig `yaml:"worker"`
	Hub    HubConfig    `yaml:"hub"`
	Front  FrontConfig  `yaml:"front"`
	Host   string       `yaml:"host" default:"127.0.0.1"`
}

type DockerConfig struct {
	Registry string `yaml:"registry" default:"docker.io/kubeshark"`
	Tag      string `yaml:"tag" default:"latest"`
}

type ResourcesConfig struct {
	Worker models.Resources `yaml:"worker"`
	Hub    models.Resources `yaml:"hub"`
}

type TapConfig struct {
	Docker                DockerConfig    `yaml:"docker"`
	Proxy                 ProxyConfig     `yaml:"proxy"`
	PodRegexStr           string          `yaml:"regex" default:".*"`
	Namespaces            []string        `yaml:"namespaces"`
	AllNamespaces         bool            `yaml:"all-namespaces" default:"false"`
	HumanMaxEntriesDBSize string          `yaml:"max-entries-db-size" default:"200MB"`
	DryRun                bool            `yaml:"dry-run" default:"false"`
	Pcap                  string          `yaml:"pcap" default:""`
	Resources             ResourcesConfig `yaml:"resources"`
	ServiceMesh           bool            `yaml:"service-mesh" default:"true"`
	Tls                   bool            `yaml:"tls" default:"true"`
	PacketCapture         string          `yaml:"packet-capture" default:"libpcap"`
	Debug                 bool            `yaml:"debug" default:"false"`
}

func (config *TapConfig) PodRegex() *regexp.Regexp {
	podRegex, _ := regexp.Compile(config.PodRegexStr)
	return podRegex
}

func (config *TapConfig) MaxEntriesDBSizeBytes() int64 {
	maxEntriesDBSizeBytes, _ := utils.HumanReadableToBytes(config.HumanMaxEntriesDBSize)
	return maxEntriesDBSizeBytes
}

func (config *TapConfig) Validate() error {
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
