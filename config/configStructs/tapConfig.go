package configStructs

import (
	"fmt"
	"regexp"

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
	SelfNamespaceLabel     = "selfnamespace"
	StorageLimitLabel      = "storagelimit"
	StorageClassLabel      = "storageclass"
	DryRunLabel            = "dryrun"
	PcapLabel              = "pcap"
	ServiceMeshLabel       = "servicemesh"
	TlsLabel               = "tls"
	IgnoreTaintedLabel     = "ignoreTainted"
	DebugLabel             = "debug"
)

type ResourceLimits struct {
	CPU    string `yaml:"cpu" default:"750m"`
	Memory string `yaml:"memory" default:"1Gi"`
}

type ResourceRequests struct {
	CPU    string `yaml:"cpu" default:"50m"`
	Memory string `yaml:"memory" default:"50Mi"`
}

type ResourceRequirements struct {
	Limits   ResourceLimits   `json:"limits"`
	Requests ResourceRequests `json:"requests"`
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
	Worker ResourceRequirements `yaml:"worker"`
	Hub    ResourceRequirements `yaml:"hub"`
}

type TapConfig struct {
	Docker            DockerConfig          `yaml:"docker"`
	Proxy             ProxyConfig           `yaml:"proxy"`
	PodRegexStr       string                `yaml:"regex" default:".*"`
	Namespaces        []string              `yaml:"namespaces"`
	SelfNamespace     string                `yaml:"selfnamespace" default:"kubeshark"`
	StorageLimit      string                `yaml:"storagelimit" default:"200Mi"`
	StorageClass      string                `yaml:"storageclass" default:"standard"`
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

func (config *TapConfig) Validate() error {
	_, compileErr := regexp.Compile(config.PodRegexStr)
	if compileErr != nil {
		return fmt.Errorf("%s is not a valid regex %s", config.PodRegexStr, compileErr)
	}

	return nil
}
