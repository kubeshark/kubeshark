package configStructs

import (
	"fmt"
	"regexp"

	v1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
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
	PersistentStorageLabel = "persistentstorage"
	StorageLimitLabel      = "storagelimit"
	StorageClassLabel      = "storageclass"
	DryRunLabel            = "dryrun"
	PcapLabel              = "pcap"
	ServiceMeshLabel       = "servicemesh"
	TlsLabel               = "tls"
	IgnoreTaintedLabel     = "ignoreTainted"
	IngressEnabledLabel    = "ingress-enabled"
	DebugLabel             = "debug"
	ContainerPort          = 80
	ContainerPortStr       = "80"
)

type ResourceLimits struct {
	CPU    string `yaml:"cpu" json:"cpu" default:"750m"`
	Memory string `yaml:"memory" json:"memory" default:"1Gi"`
}

type ResourceRequests struct {
	CPU    string `yaml:"cpu" json:"cpu" default:"50m"`
	Memory string `yaml:"memory" json:"memory" default:"50Mi"`
}

type ResourceRequirements struct {
	Limits   ResourceLimits   `yaml:"limits" json:"limits"`
	Requests ResourceRequests `yaml:"requests" json:"requests"`
}

type WorkerConfig struct {
	SrvPort uint16 `yaml:"srvport" json:"srvport" default:"8897"`
}

type HubConfig struct {
	Port    uint16 `yaml:"port" json:"port" default:"8898"`
	SrvPort uint16 `yaml:"srvport" json:"srvport" default:"8898"`
}

type FrontConfig struct {
	Port    uint16 `yaml:"port" json:"port" default:"8899"`
	SrvPort uint16 `yaml:"srvport" json:"srvport" default:"8899"`
}

type ProxyConfig struct {
	Worker WorkerConfig `yaml:"worker" json:"worker"`
	Hub    HubConfig    `yaml:"hub" json:"hub"`
	Front  FrontConfig  `yaml:"front" json:"front"`
	Host   string       `yaml:"host" json:"host" default:"127.0.0.1"`
}

type DockerConfig struct {
	Registry         string   `yaml:"registry" json:"registry" default:"docker.io/kubeshark"`
	Tag              string   `yaml:"tag" json:"tag" default:"latest"`
	ImagePullPolicy  string   `yaml:"imagepullpolicy" json:"imagepullpolicy" default:"Always"`
	ImagePullSecrets []string `yaml:"imagepullsecrets" json:"imagepullsecrets"`
}

type ResourcesConfig struct {
	Worker ResourceRequirements `yaml:"worker"`
	Hub    ResourceRequirements `yaml:"hub"`
}

type AuthConfig struct {
	ApprovedDomains []string `yaml:"approvedDomains" json:"approvedDomains"`
}

type IngressConfig struct {
	Enabled     bool                    `yaml:"enabled" json:"enabled" default:"false"`
	Host        string                  `yaml:"host" json:"host" default:"ks.svc.cluster.local"`
	TLS         []networking.IngressTLS `yaml:"tls" json:"tls"`
	Auth        AuthConfig              `yaml:"auth" json:"auth"`
	CertManager string                  `yaml:"certManager" json:"certManager" default:"letsencrypt-prod"`
}

type TapConfig struct {
	Docker            DockerConfig          `yaml:"docker" json:"docker"`
	Proxy             ProxyConfig           `yaml:"proxy" json:"proxy"`
	PodRegexStr       string                `yaml:"regex" json:"regex" default:".*"`
	Namespaces        []string              `yaml:"namespaces" json:"namespaces"`
	SelfNamespace     string                `yaml:"selfnamespace" json:"selfnamespace" default:"kubeshark"`
	PersistentStorage bool                  `yaml:"persistentstorage" json:"persistentstorage" default:"false"`
	StorageLimit      string                `yaml:"storagelimit" json:"storagelimit" default:"200Mi"`
	StorageClass      string                `yaml:"storageclass" json:"storageclass" default:"standard"`
	DryRun            bool                  `yaml:"dryrun" json:"dryrun" default:"false"`
	Pcap              string                `yaml:"pcap" json:"pcap" default:""`
	Resources         ResourcesConfig       `yaml:"resources" json:"resources"`
	ServiceMesh       bool                  `yaml:"servicemesh" json:"servicemesh" default:"true"`
	Tls               bool                  `yaml:"tls" json:"tls" default:"true"`
	PacketCapture     string                `yaml:"packetcapture" json:"packetcapture" default:"libpcap"`
	IgnoreTainted     bool                  `yaml:"ignoreTainted" json:"ignoreTainted" default:"false"`
	ResourceLabels    map[string]string     `yaml:"resourceLabels" json:"resourceLabels" default:"{}"`
	NodeSelectorTerms []v1.NodeSelectorTerm `yaml:"nodeSelectorTerms" json:"nodeSelectorTerms" default:"[]"`
	Ingress           IngressConfig         `yaml:"ingress" json:"ingress"`
	Debug             bool                  `yaml:"debug" json:"debug" default:"false"`
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
