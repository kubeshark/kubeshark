package configStructs

import (
	"fmt"
	"regexp"

	v1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
)

const (
	DockerRegistryLabel          = "docker-registry"
	DockerTagLabel               = "docker-tag"
	DockerImagePullPolicy        = "docker-imagePullPolicy"
	DockerImagePullSecrets       = "docker-imagePullSecrets"
	ProxyFrontPortLabel          = "proxy-front-port"
	ProxyHubPortLabel            = "proxy-hub-port"
	ProxyHostLabel               = "proxy-host"
	NamespacesLabel              = "namespaces"
	ExcludedNamespacesLabel      = "excludedNamespaces"
	ReleaseNamespaceLabel        = "release-namespace"
	PersistentStorageLabel       = "persistentStorage"
	PersistentStorageStaticLabel = "persistentStorageStatic"
	EfsFileSytemIdAndPathLabel   = "efsFileSytemIdAndPath"
	StorageLimitLabel            = "storageLimit"
	StorageClassLabel            = "storageClass"
	DryRunLabel                  = "dryRun"
	PcapLabel                    = "pcap"
	ServiceMeshLabel             = "serviceMesh"
	TlsLabel                     = "tls"
	IgnoreTaintedLabel           = "ignoreTainted"
	IngressEnabledLabel          = "ingress-enabled"
	TelemetryEnabledLabel        = "telemetry-enabled"
	PprofPortLabel               = "pprof-port"
	PprofViewLabel               = "pprof-view"
	DebugLabel                   = "debug"
	ContainerPort                = 80
	ContainerPortStr             = "80"
)

type ResourceLimitsHub struct {
	CPU    string `yaml:"cpu" json:"cpu" default:"1000m"`
	Memory string `yaml:"memory" json:"memory" default:"1500Mi"`
}

type ResourceLimitsWorker struct {
	CPU    string `yaml:"cpu" json:"cpu" default:"1200m"`
	Memory string `yaml:"memory" json:"memory" default:"2000Mi"`
}

type ResourceRequests struct {
	CPU    string `yaml:"cpu" json:"cpu" default:"50m"`
	Memory string `yaml:"memory" json:"memory" default:"50Mi"`
}

type ResourceRequirementsHub struct {
	Limits   ResourceLimitsHub `yaml:"limits" json:"limits"`
	Requests ResourceRequests  `yaml:"requests" json:"requests"`
}

type ResourceRequirementsWorker struct {
	Limits   ResourceLimitsHub `yaml:"limits" json:"limits"`
	Requests ResourceRequests  `yaml:"requests" json:"requests"`
}

type WorkerConfig struct {
	SrvPort uint16 `yaml:"srvPort" json:"srvPort" default:"30001"`
}

type HubConfig struct {
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

type OverrideTagConfig struct {
	Worker string `yaml:"worker" json:"worker"`
	Hub    string `yaml:"hub" json:"hub"`
	Front  string `yaml:"front" json:"front"`
}

type DockerConfig struct {
	Registry         string            `yaml:"registry" json:"registry" default:"docker.io/kubeshark"`
	Tag              string            `yaml:"tag" json:"tag" default:""`
	TagLocked        bool              `yaml:"tagLocked" json:"tagLocked" default:"true"`
	ImagePullPolicy  string            `yaml:"imagePullPolicy" json:"imagePullPolicy" default:"Always"`
	ImagePullSecrets []string          `yaml:"imagePullSecrets" json:"imagePullSecrets"`
	OverrideTag      OverrideTagConfig `yaml:"overrideTag" json:"overrideTag"`
}

type ResourcesConfig struct {
	Hub     ResourceRequirementsHub    `yaml:"hub" json:"hub"`
	Sniffer ResourceRequirementsWorker `yaml:"sniffer" json:"sniffer"`
	Tracer  ResourceRequirementsWorker `yaml:"tracer" json:"tracer"`
}

type Role struct {
	Filter                  string `yaml:"filter" json:"filter" default:""`
	CanDownloadPCAP         bool   `yaml:"canDownloadPCAP" json:"canDownloadPCAP" default:"false"`
	CanUseScripting         bool   `yaml:"canUseScripting" json:"canUseScripting" default:"false"`
	CanUpdateTargetedPods   bool   `yaml:"canUpdateTargetedPods" json:"canUpdateTargetedPods" default:"false"`
	CanStopTrafficCapturing bool   `yaml:"canStopTrafficCapturing" json:"canStopTrafficCapturing" default:"false"`
	ShowAdminConsoleLink    bool   `yaml:"showAdminConsoleLink" json:"showAdminConsoleLink" default:"false"`
}

type SamlConfig struct {
	IdpMetadataUrl string          `yaml:"idpMetadataUrl" json:"idpMetadataUrl"`
	X509crt        string          `yaml:"x509crt" json:"x509crt"`
	X509key        string          `yaml:"x509key" json:"x509key"`
	RoleAttribute  string          `yaml:"roleAttribute" json:"roleAttribute"`
	Roles          map[string]Role `yaml:"roles" json:"roles"`
}

type AuthConfig struct {
	Enabled bool       `yaml:"enabled" json:"enabled" default:"false"`
	Type    string     `yaml:"type" json:"type" default:"saml"`
	Saml    SamlConfig `yaml:"saml" json:"saml"`
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

type SentryConfig struct {
	Enabled     bool   `yaml:"enabled" json:"enabled" default:"false"`
	Environment string `yaml:"environment" json:"environment" default:"production"`
}

type CapabilitiesConfig struct {
	NetworkCapture     []string `yaml:"networkCapture" json:"networkCapture"  default:"[]"`
	ServiceMeshCapture []string `yaml:"serviceMeshCapture" json:"serviceMeshCapture"  default:"[]"`
	KernelModule       []string `yaml:"kernelModule" json:"kernelModule"  default:"[]"`
	EBPFCapture        []string `yaml:"ebpfCapture" json:"ebpfCapture"  default:"[]"`
}

type KernelModuleConfig struct {
	Enabled         bool   `yaml:"enabled" json:"enabled" default:"false"`
	Image           string `yaml:"image" json:"image" default:"kubeshark/pf-ring-module:all"`
	UnloadOnDestroy bool   `yaml:"unloadOnDestroy" json:"unloadOnDestroy" default:"false"`
}

type MetricsConfig struct {
	Port uint16 `yaml:"port" json:"port" default:"49100"`
}

type PprofConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled" default:"false"`
	Port    uint16 `yaml:"port" json:"port" default:"8000"`
	View    string `yaml:"view" json:"view" default:"flamegraph"`
}

type MiscConfig struct {
	JsonTTL                     string `yaml:"jsonTTL" json:"jsonTTL" default:"5m"`
	PcapTTL                     string `yaml:"pcapTTL" json:"pcapTTL" default:"10s"`
	PcapErrorTTL                string `yaml:"pcapErrorTTL" json:"pcapErrorTTL" default:"60s"`
	TrafficSampleRate           int    `yaml:"trafficSampleRate" json:"trafficSampleRate" default:"100"`
	TcpStreamChannelTimeoutMs   int    `yaml:"tcpStreamChannelTimeoutMs" json:"tcpStreamChannelTimeoutMs" default:"10000"`
	TcpStreamChannelTimeoutShow bool   `yaml:"tcpStreamChannelTimeoutShow" json:"tcpStreamChannelTimeoutShow" default:"false"`
	ResolutionStrategy          string `yaml:"resolutionStrategy" json:"resolutionStrategy" default:"auto"`
	DuplicateTimeframe          string `yaml:"duplicateTimeframe" json:"duplicateTimeframe" default:"200ms"`
	DetectDuplicates            bool   `yaml:"detectDuplicates" json:"detectDuplicates" default:"false"`
}

type TapConfig struct {
	Docker                       DockerConfig          `yaml:"docker" json:"docker"`
	Proxy                        ProxyConfig           `yaml:"proxy" json:"proxy"`
	PodRegexStr                  string                `yaml:"regex" json:"regex" default:".*"`
	Namespaces                   []string              `yaml:"namespaces" json:"namespaces" default:"[]"`
	ExcludedNamespaces           []string              `yaml:"excludedNamespaces" json:"excludedNamespaces" default:"[]"`
	BpfOverride                  string                `yaml:"bpfOverride" json:"bpfOverride" default:""`
	Stopped                      bool                  `yaml:"stopped" json:"stopped" default:"true"`
	Release                      ReleaseConfig         `yaml:"release" json:"release"`
	PersistentStorage            bool                  `yaml:"persistentStorage" json:"persistentStorage" default:"false"`
	PersistentStorageStatic      bool                  `yaml:"persistentStorageStatic" json:"persistentStorageStatic" default:"false"`
	EfsFileSytemIdAndPath        string                `yaml:"efsFileSytemIdAndPath" json:"efsFileSytemIdAndPath" default:""`
	StorageLimit                 string                `yaml:"storageLimit" json:"storageLimit" default:"5000Mi"`
	StorageClass                 string                `yaml:"storageClass" json:"storageClass" default:"standard"`
	DryRun                       bool                  `yaml:"dryRun" json:"dryRun" default:"false"`
	Resources                    ResourcesConfig       `yaml:"resources" json:"resources"`
	ServiceMesh                  bool                  `yaml:"serviceMesh" json:"serviceMesh" default:"true"`
	Tls                          bool                  `yaml:"tls" json:"tls" default:"true"`
	DisableTlsLog                bool                  `yaml:"disableTlsLog" json:"disableTlsLog" default:"false"`
	PacketCapture                string                `yaml:"packetCapture" json:"packetCapture" default:"best"`
	IgnoreTainted                bool                  `yaml:"ignoreTainted" json:"ignoreTainted" default:"false"`
	Labels                       map[string]string     `yaml:"labels" json:"labels" default:"{}"`
	Annotations                  map[string]string     `yaml:"annotations" json:"annotations" default:"{}"`
	NodeSelectorTerms            []v1.NodeSelectorTerm `yaml:"nodeSelectorTerms" json:"nodeSelectorTerms" default:"[]"`
	Auth                         AuthConfig            `yaml:"auth" json:"auth"`
	Ingress                      IngressConfig         `yaml:"ingress" json:"ingress"`
	IPv6                         bool                  `yaml:"ipv6" json:"ipv6" default:"true"`
	Debug                        bool                  `yaml:"debug" json:"debug" default:"false"`
	KernelModule                 KernelModuleConfig    `yaml:"kernelModule" json:"kernelModule"`
	Telemetry                    TelemetryConfig       `yaml:"telemetry" json:"telemetry"`
	Sentry                       SentryConfig          `yaml:"sentry" json:"sentry"`
	DefaultFilter                string                `yaml:"defaultFilter" json:"defaultFilter" default:"!dns and !tcp"`
	ScriptingDisabled            bool                  `yaml:"scriptingDisabled" json:"scriptingDisabled" default:"false"`
	TargetedPodsUpdateDisabled   bool                  `yaml:"targetedPodsUpdateDisabled" json:"targetedPodsUpdateDisabled" default:"false"`
	RecordingDisabled            bool                  `yaml:"recordingDisabled" json:"recordingDisabled" default:"false"`
	StopTrafficCapturingDisabled bool                  `yaml:"stopTrafficCapturingDisabled" json:"stopTrafficCapturingDisabled" default:"false"`
	Capabilities                 CapabilitiesConfig    `yaml:"capabilities" json:"capabilities"`
	GlobalFilter                 string                `yaml:"globalFilter" json:"globalFilter"`
	EnabledDissectors            []string              `yaml:"enabledDissectors" json:"enabledDissectors"`
	Metrics                      MetricsConfig         `yaml:"metrics" json:"metrics"`
	Pprof                        PprofConfig           `yaml:"pprof" json:"pprof"`
	Misc                         MiscConfig            `yaml:"misc" json:"misc"`
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
