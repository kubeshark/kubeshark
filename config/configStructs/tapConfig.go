package configStructs

import (
	"fmt"
	"regexp"

	"github.com/kubeshark/kubeshark/utils"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

const (
	DockerRegistryLabel    = "docker-registry"
	DockerTagLabel         = "docker-tag"
	DockerImagePullPolicy  = "docker-imagepullpolicy"
	DockerImagePullSecrets = "docker-imagepullsecrets"
	ProxyFrontPortLabel    = "proxy-front-port"
	ProxyHubPortLabel      = "proxy-hub-port"
	ProxyHostLabel         = "proxy-host"
	NamespacesLabel        = "namespaces"
	AllNamespacesLabel     = "allnamespaces"
	SelfNamespaceLabel     = "selfnamespace"
	StorageLimitLabel      = "storagelimit"
	DryRunLabel            = "dryrun"
	PcapLabel              = "pcap"
	ServiceMeshLabel       = "servicemesh"
	TlsLabel               = "tls"
	IgnoreTaintedLabel     = "ignoreTainted"
	DebugLabel             = "debug"
)

type Resources struct {
	CpuLimit       string `yaml:"cpu-limit" default:"750m"`
	MemoryLimit    string `yaml:"memory-limit" default:"1Gi"`
	CpuRequests    string `yaml:"cpu-requests" default:"50m"`
	MemoryRequests string `yaml:"memory-requests" default:"50Mi"`
}

type WorkerConfig struct {
	SrcPort uint16 `yaml:"port" default:"8897"`
	DstPort uint16 `yaml:"srvport" default:"8897"`
}

type HubConfig struct {
	SrcPort uint16 `yaml:"port" default:"8898"`
	DstPort uint16 `yaml:"srvport" default:"80"`
}

type FrontConfig struct {
	SrcPort uint16 `yaml:"port" default:"8899"`
	DstPort uint16 `yaml:"srvport" default:"80"`
}

type ProxyConfig struct {
	Worker WorkerConfig `yaml:"worker"`
	Hub    HubConfig    `yaml:"hub"`
	Front  FrontConfig  `yaml:"front"`
	Host   string       `yaml:"host" default:"127.0.0.1"`
}

type DockerConfig struct {
	Registry         string   `yaml:"registry" default:"docker.io/kubeshark"`
	Tag              string   `yaml:"tag" default:"latest"`
	ImagePullPolicy  string   `yaml:"imagepullpolicy" default:"Always"`
	ImagePullSecrets []string `yaml:"imagepullsecrets"`
}

type ResourcesConfig struct {
	Worker Resources `yaml:"worker"`
	Hub    Resources `yaml:"hub"`
}

type TapConfig struct {
	Docker            DockerConfig          `yaml:"docker"`
	Proxy             ProxyConfig           `yaml:"proxy"`
	PodRegexStr       string                `yaml:"regex" default:".*"`
	Namespaces        []string              `yaml:"namespaces"`
	AllNamespaces     bool                  `yaml:"allnamespaces" default:"true"`
	SelfNamespace     string                `yaml:"selfnamespace" default:"kubeshark"`
	StorageLimit      string                `yaml:"storagelimit" default:"200MB"`
	DryRun            bool                  `yaml:"dryrun" default:"false"`
	Pcap              string                `yaml:"pcap" default:""`
	Resources         ResourcesConfig       `yaml:"resources"`
	ServiceMesh       bool                  `yaml:"servicemesh" default:"true"`
	Tls               bool                  `yaml:"tls" default:"true"`
	PacketCapture     string                `yaml:"packetcapture" default:"libpcap"`
	IgnoreTainted     bool                  `yaml:"ignoreTainted" default:"false"`
	ResourceLabels    map[string]string     `yaml:"resourceLabels" default:"{}"`
	NodeSelectorTerms []v1.NodeSelectorTerm `yaml:"nodeSelectorTerms" default:"[]"`
	Debug             bool                  `yaml:"debug" default:"false"`
}

func (config *TapConfig) PodRegex() *regexp.Regexp {
	podRegex, _ := regexp.Compile(config.PodRegexStr)
	return podRegex
}

func (config *TapConfig) StorageLimitBytes() int64 {
	storageLimitBytes, err := utils.HumanReadableToBytes(config.StorageLimit)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	return storageLimitBytes
}

func (config *TapConfig) Validate() error {
	_, compileErr := regexp.Compile(config.PodRegexStr)
	if compileErr != nil {
		return fmt.Errorf("%s is not a valid regex %s", config.PodRegexStr, compileErr)
	}

	_, parseHumanDataSizeErr := utils.HumanReadableToBytes(config.StorageLimit)
	if parseHumanDataSizeErr != nil {
		return fmt.Errorf("Could not parse --%s value %s", StorageLimitLabel, config.StorageLimit)
	}

	return nil
}
