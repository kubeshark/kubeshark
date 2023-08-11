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
	ReleaseNamespaceLabel  = "release-namespace"
	PersistentStorageLabel = "persistentstorage"
	StorageLimitLabel      = "storagelimit"
	StorageClassLabel      = "storageclass"
	DryRunLabel            = "dryrun"
	PcapLabel              = "pcap"
	ServiceMeshLabel       = "servicemesh"
	TlsLabel               = "tls"
	IgnoreTaintedLabel     = "ignoretainted"
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
	Port uint16 `yaml:"port" json:"port" default:"8899"`
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
	Worker ResourceRequirements `yaml:"worker" json:"worker"`
	Hub    ResourceRequirements `yaml:"hub" json:"hub"`
}

type AuthConfig struct {
	Enabled         bool     `yaml:"enabled" json:"enabled" default:"false"`
	ApprovedEmails  []string `yaml:"approvedemails" json:"approvedemails"  default:"[]"`
	ApprovedDomains []string `yaml:"approveddomains" json:"approveddomains"  default:"[]"`
}

type IngressConfig struct {
	Enabled     bool                    `yaml:"enabled" json:"enabled" default:"false"`
	ClassName   string                  `yaml:"classname" json:"classname" default:"kubeshark-ingress-class"`
	Controller  string                  `yaml:"controller" json:"controller" default:"k8s.io/ingress-nginx"`
	Host        string                  `yaml:"host" json:"host" default:"ks.svc.cluster.local"`
	TLS         []networking.IngressTLS `yaml:"tls" json:"tls"`
	CertManager string                  `yaml:"certmanager" json:"certmanager" default:"letsencrypt-prod"`
}

type ReleaseConfig struct {
	Repo      string `yaml:"repo" json:"repo" default:"https://helm.kubeshark.co"`
	Name      string `yaml:"name" json:"name" default:"kubeshark"`
	Namespace string `yaml:"namespace" json:"namespace" default:"default"`
}

type KMMConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled" default:"true"`
}

type TapConfig struct {
	Docker            DockerConfig          `yaml:"docker" json:"docker"`
	Proxy             ProxyConfig           `yaml:"proxy" json:"proxy"`
	PodRegexStr       string                `yaml:"regex" json:"regex" default:".*"`
	Namespaces        []string              `yaml:"namespaces" json:"namespaces" default:"[]"`
	Release           ReleaseConfig         `yaml:"release" json:"release"`
	PersistentStorage bool                  `yaml:"persistentstorage" json:"persistentstorage" default:"false"`
	StorageLimit      string                `yaml:"storagelimit" json:"storagelimit" default:"200Mi"`
	StorageClass      string                `yaml:"storageclass" json:"storageclass" default:"standard"`
	DryRun            bool                  `yaml:"dryrun" json:"dryrun" default:"false"`
	Pcap              string                `yaml:"pcap" json:"pcap" default:""`
	Resources         ResourcesConfig       `yaml:"resources" json:"resources"`
	ServiceMesh       bool                  `yaml:"servicemesh" json:"servicemesh" default:"true"`
	Tls               bool                  `yaml:"tls" json:"tls" default:"true"`
	PacketCapture     string                `yaml:"packetcapture" json:"packetcapture" default:"libpcap"`
	IgnoreTainted     bool                  `yaml:"ignoretainted" json:"ignoretainted" default:"false"`
	Labels            map[string]string     `yaml:"labels" json:"labels" default:"{}"`
	Annotations       map[string]string     `yaml:"annotations" json:"annotations" default:"{}"`
	NodeSelectorTerms []v1.NodeSelectorTerm `yaml:"nodeselectorterms" json:"nodeselectorterms" default:"[]"`
	Auth              AuthConfig            `yaml:"auth" json:"auth"`
	Ingress           IngressConfig         `yaml:"ingress" json:"ingress"`
	KMM               KMMConfig             `yaml:"kmm" json:"kmm"`
	IPv6              bool                  `yaml:"ipv6" json:"ipv6" default:"true"`
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
