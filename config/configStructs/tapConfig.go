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
	DockerImagePullPolicy  = "docker-imagePullPolicy"
	DockerImagePullSecrets = "docker-imagePullSecrets"
	ProxyFrontPortLabel    = "proxy-front-port"
	ProxyHubPortLabel      = "proxy-hub-port"
	ProxyHostLabel         = "proxy-host"
	NamespacesLabel        = "namespaces"
	ReleaseNamespaceLabel  = "release-namespace"
	PersistentStorageLabel = "persistentStorage"
	StorageLimitLabel      = "storageLimit"
	StorageClassLabel      = "storageClass"
	DryRunLabel            = "dryRun"
	PcapLabel              = "pcap"
	ServiceMeshLabel       = "serviceMesh"
	TlsLabel               = "tls"
	IgnoreTaintedLabel     = "ignoreTainted"
	IngressEnabledLabel    = "ingress-enabled"
	TelemetryEnabledLabel  = "telemetry-enabled"
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
	SrvPort uint16 `yaml:"srvPort" json:"srvPort" default:"8897"`
}

type HubConfig struct {
	Port    uint16 `yaml:"port" json:"port" default:"8898"`
	SrvPort uint16 `yaml:"srvPort" json:"srvPort" default:"8898"`
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
	Tag              string   `yaml:"tag" json:"tag" default:""`
	ImagePullPolicy  string   `yaml:"imagePullPolicy" json:"imagePullPolicy" default:"Always"`
	ImagePullSecrets []string `yaml:"imagePullSecrets" json:"imagePullSecrets"`
}

type ResourcesConfig struct {
	Worker ResourceRequirements `yaml:"worker" json:"worker"`
	Hub    ResourceRequirements `yaml:"hub" json:"hub"`
}

type AuthConfig struct {
	Enabled         bool     `yaml:"enabled" json:"enabled" default:"false"`
	ApprovedEmails  []string `yaml:"approvedEmails" json:"approvedEmails"  default:"[]"`
	ApprovedDomains []string `yaml:"approvedDomains" json:"approvedDomains"  default:"[]"`
	ApprovedTenants []string `yaml:"approvedTenants" json:"approvedTenants"  default:"[]"`
}

type IngressConfig struct {
	Enabled     bool                    `yaml:"enabled" json:"enabled" default:"false"`
	ClassName   string                  `yaml:"className" json:"className" default:""`
	Host        string                  `yaml:"host" json:"host" default:"ks.svc.cluster.local"`
	TLS         []networking.IngressTLS `yaml:"tls" json:"tls" default:"[]"`
	Annotations map[string]string       `yaml:"annotations" json:"annotations" default:"{}"`
}

type ReleaseConfig struct {
	Repo      string `yaml:"repo" json:"repo" default:"https://helm.kubeshark.co"`
	Name      string `yaml:"name" json:"name" default:"kubeshark"`
	Namespace string `yaml:"namespace" json:"namespace" default:"default"`
}

type TelemetryConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled" default:"true"`
}

type TapConfig struct {
	Docker            DockerConfig          `yaml:"docker" json:"docker"`
	Proxy             ProxyConfig           `yaml:"proxy" json:"proxy"`
	PodRegexStr       string                `yaml:"regex" json:"regex" default:".*"`
	Namespaces        []string              `yaml:"namespaces" json:"namespaces" default:"[]"`
	Release           ReleaseConfig         `yaml:"release" json:"release"`
	PersistentStorage bool                  `yaml:"persistentStorage" json:"persistentStorage" default:"false"`
	StorageLimit      string                `yaml:"storageLimit" json:"storageLimit" default:"500Mi"`
	StorageClass      string                `yaml:"storageClass" json:"storageClass" default:"standard"`
	DryRun            bool                  `yaml:"dryRun" json:"dryRun" default:"false"`
	Pcap              string                `yaml:"pcap" json:"pcap" default:""`
	Resources         ResourcesConfig       `yaml:"resources" json:"resources"`
	ServiceMesh       bool                  `yaml:"serviceMesh" json:"serviceMesh" default:"true"`
	Tls               bool                  `yaml:"tls" json:"tls" default:"true"`
	IgnoreTainted     bool                  `yaml:"ignoreTainted" json:"ignoreTainted" default:"false"`
	Labels            map[string]string     `yaml:"labels" json:"labels" default:"{}"`
	Annotations       map[string]string     `yaml:"annotations" json:"annotations" default:"{}"`
	NodeSelectorTerms []v1.NodeSelectorTerm `yaml:"nodeSelectorTerms" json:"nodeSelectorTerms" default:"[]"`
	Auth              AuthConfig            `yaml:"auth" json:"auth"`
	Ingress           IngressConfig         `yaml:"ingress" json:"ingress"`
	IPv6              bool                  `yaml:"ipv6" json:"ipv6" default:"true"`
	Debug             bool                  `yaml:"debug" json:"debug" default:"false"`
	NoKernelModule    bool                  `yaml:"noKernelModule" json:"noKernelModule" default:"false"`
	Telemetry         TelemetryConfig       `yaml:"telemetry" json:"telemetry"`
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
