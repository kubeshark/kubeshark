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
	ResourceGuardEnabledLabel    = "resource-guard-enabled"
	PprofPortLabel               = "pprof-port"
	PprofViewLabel               = "pprof-view"
	DebugLabel                   = "debug"
	ContainerPort                = 8080
	ContainerPortStr             = "8080"
	PcapDest                     = "dest"
	PcapMaxSize                  = "maxSize"
	PcapMaxTime                  = "maxTime"
	PcapTimeInterval             = "timeInterval"
	PcapKubeconfig               = "kubeconfig"
	PcapDumpEnabled              = "enabled"
	PcapTime                     = "time"
	WatchdogEnabled              = "watchdogEnabled"
)

type ResourceLimitsHub struct {
	CPU    string `yaml:"cpu" json:"cpu" default:"0"`
	Memory string `yaml:"memory" json:"memory" default:"5Gi"`
}

type ResourceLimitsWorker struct {
	CPU    string `yaml:"cpu" json:"cpu" default:"0"`
	Memory string `yaml:"memory" json:"memory" default:"3Gi"`
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
	SrvPort uint16 `yaml:"srvPort" json:"srvPort" default:"48999"`
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

type OverrideImageConfig struct {
	Worker string `yaml:"worker" json:"worker"`
	Hub    string `yaml:"hub" json:"hub"`
	Front  string `yaml:"front" json:"front"`
}
type OverrideTagConfig struct {
	Worker string `yaml:"worker" json:"worker"`
	Hub    string `yaml:"hub" json:"hub"`
	Front  string `yaml:"front" json:"front"`
}

type DockerConfig struct {
	Registry         string              `yaml:"registry" json:"registry" default:"docker.io/kubeshark"`
	Tag              string              `yaml:"tag" json:"tag" default:""`
	TagLocked        bool                `yaml:"tagLocked" json:"tagLocked" default:"true"`
	ImagePullPolicy  string              `yaml:"imagePullPolicy" json:"imagePullPolicy" default:"Always"`
	ImagePullSecrets []string            `yaml:"imagePullSecrets" json:"imagePullSecrets"`
	OverrideImage    OverrideImageConfig `yaml:"overrideImage" json:"overrideImage"`
	OverrideTag      OverrideTagConfig   `yaml:"overrideTag" json:"overrideTag"`
}

type DnsConfig struct {
	Nameservers []string          `yaml:"nameservers" json:"nameservers" default:"[]"`
	Searches    []string          `yaml:"searches" json:"searches" default:"[]"`
	Options     []DnsConfigOption `yaml:"options" json:"options" default:"[]"`
}

type DnsConfigOption struct {
	Name  string `yaml:"name" json:"name"`
	Value string `yaml:"value" json:"value"`
}

type ResourcesConfig struct {
	Hub     ResourceRequirementsHub    `yaml:"hub" json:"hub"`
	Sniffer ResourceRequirementsWorker `yaml:"sniffer" json:"sniffer"`
	Tracer  ResourceRequirementsWorker `yaml:"tracer" json:"tracer"`
}

type ProbesConfig struct {
	Hub     ProbeConfig `yaml:"hub" json:"hub"`
	Sniffer ProbeConfig `yaml:"sniffer" json:"sniffer"`
}

type NodeSelectorTermsConfig struct {
	Hub     []v1.NodeSelectorTerm `yaml:"hub" json:"hub" default:"[]"`
	Workers []v1.NodeSelectorTerm `yaml:"workers" json:"workers" default:"[]"`
	Front   []v1.NodeSelectorTerm `yaml:"front" json:"front" default:"[]"`
	Dex     []v1.NodeSelectorTerm `yaml:"dex" json:"dex" default:"[]"`
}

type TolerationsConfig struct {
	Hub     []v1.Toleration `yaml:"hub" json:"hub" default:"[]"`
	Workers []v1.Toleration `yaml:"workers" json:"workers" default:"[]"`
	Front   []v1.Toleration `yaml:"front" json:"front" default:"[]"`
}

type ProbeConfig struct {
	InitialDelaySeconds int `yaml:"initialDelaySeconds" json:"initialDelaySeconds" default:"5"`
	PeriodSeconds       int `yaml:"periodSeconds" json:"periodSeconds" default:"5"`
	SuccessThreshold    int `yaml:"successThreshold" json:"successThreshold" default:"1"`
	FailureThreshold    int `yaml:"failureThreshold" json:"failureThreshold" default:"3"`
}

type ScriptingPermissions struct {
	CanSave     bool `yaml:"canSave" json:"canSave" default:"true"`
	CanActivate bool `yaml:"canActivate" json:"canActivate" default:"true"`
	CanDelete   bool `yaml:"canDelete" json:"canDelete" default:"true"`
}

type Role struct {
	Filter                  string               `yaml:"filter" json:"filter" default:""`
	CanDownloadPCAP         bool                 `yaml:"canDownloadPCAP" json:"canDownloadPCAP" default:"false"`
	CanUseScripting         bool                 `yaml:"canUseScripting" json:"canUseScripting" default:"false"`
	ScriptingPermissions    ScriptingPermissions `yaml:"scriptingPermissions" json:"scriptingPermissions"`
	CanUpdateTargetedPods   bool                 `yaml:"canUpdateTargetedPods" json:"canUpdateTargetedPods" default:"false"`
	CanStopTrafficCapturing bool                 `yaml:"canStopTrafficCapturing" json:"canStopTrafficCapturing" default:"false"`
	ShowAdminConsoleLink    bool                 `yaml:"showAdminConsoleLink" json:"showAdminConsoleLink" default:"false"`
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

type RoutingConfig struct {
	Front FrontRoutingConfig `yaml:"front" json:"front"`
}

type DashboardConfig struct {
	CompleteStreamingEnabled bool `yaml:"completeStreamingEnabled" json:"completeStreamingEnabled" default:"true"`
}

type FrontRoutingConfig struct {
	BasePath string `yaml:"basePath" json:"basePath" default:""`
}

type ReleaseConfig struct {
	Repo      string `yaml:"repo" json:"repo" default:"https://helm.kubeshark.co"`
	Name      string `yaml:"name" json:"name" default:"kubeshark"`
	Namespace string `yaml:"namespace" json:"namespace" default:"default"`
}

type TelemetryConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled" default:"true"`
}

type ResourceGuardConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled" default:"false"`
}

type SentryConfig struct {
	Enabled     bool   `yaml:"enabled" json:"enabled" default:"false"`
	Environment string `yaml:"environment" json:"environment" default:"production"`
}

type WatchdogConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled" default:"false"`
}

type GitopsConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled" default:"false"`
}

type CapabilitiesConfig struct {
	NetworkCapture     []string `yaml:"networkCapture" json:"networkCapture"  default:"[]"`
	ServiceMeshCapture []string `yaml:"serviceMeshCapture" json:"serviceMeshCapture"  default:"[]"`
	EBPFCapture        []string `yaml:"ebpfCapture" json:"ebpfCapture"  default:"[]"`
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
	StaleTimeoutSeconds         int    `yaml:"staleTimeoutSeconds" json:"staleTimeoutSeconds" default:"30"`
}

type PcapDumpConfig struct {
	PcapDumpEnabled  bool   `yaml:"enabled" json:"enabled" default:"true"`
	PcapTimeInterval string `yaml:"timeInterval" json:"timeInterval" default:"1m"`
	PcapMaxTime      string `yaml:"maxTime" json:"maxTime" default:"1h"`
	PcapMaxSize      string `yaml:"maxSize" json:"maxSize" default:"500MB"`
	PcapTime         string `yaml:"time" json:"time" default:"time"`
	PcapDebug        bool   `yaml:"debug" json:"debug" default:"false"`
	PcapDest         string `yaml:"dest" json:"dest" default:""`
}

type PortMapping struct {
	HTTP     []uint16 `yaml:"http" json:"http"`
	AMQP     []uint16 `yaml:"amqp" json:"amqp"`
	KAFKA    []uint16 `yaml:"kafka" json:"kafka"`
	REDIS    []uint16 `yaml:"redis" json:"redis"`
	LDAP     []uint16 `yaml:"ldap" json:"ldap"`
	DIAMETER []uint16 `yaml:"diameter" json:"diameter"`
}

type SecurityContextConfig struct {
	Privileged      bool                  `yaml:"privileged" json:"privileged" default:"true"`
	AppArmorProfile AppArmorProfileConfig `yaml:"appArmorProfile" json:"appArmorProfile"`
	SeLinuxOptions  SeLinuxOptionsConfig  `yaml:"seLinuxOptions" json:"seLinuxOptions"`
	Capabilities    CapabilitiesConfig    `yaml:"capabilities" json:"capabilities"`
}

type AppArmorProfileConfig struct {
	Type             string `yaml:"type" json:"type"`
	LocalhostProfile string `yaml:"localhostProfile" json:"localhostProfile"`
}

type SeLinuxOptionsConfig struct {
	Level string `yaml:"level" json:"level"`
	Role  string `yaml:"role" json:"role"`
	Type  string `yaml:"type" json:"type"`
	User  string `yaml:"user" json:"user"`
}

type TapConfig struct {
	Docker                         DockerConfig            `yaml:"docker" json:"docker"`
	Proxy                          ProxyConfig             `yaml:"proxy" json:"proxy"`
	PodRegexStr                    string                  `yaml:"regex" json:"regex" default:".*"`
	Namespaces                     []string                `yaml:"namespaces" json:"namespaces" default:"[]"`
	ExcludedNamespaces             []string                `yaml:"excludedNamespaces" json:"excludedNamespaces" default:"[]"`
	BpfOverride                    string                  `yaml:"bpfOverride" json:"bpfOverride" default:""`
	Stopped                        bool                    `yaml:"stopped" json:"stopped" default:"false"`
	Release                        ReleaseConfig           `yaml:"release" json:"release"`
	PersistentStorage              bool                    `yaml:"persistentStorage" json:"persistentStorage" default:"false"`
	PersistentStorageStatic        bool                    `yaml:"persistentStorageStatic" json:"persistentStorageStatic" default:"false"`
	PersistentStoragePvcVolumeMode string                  `yaml:"persistentStoragePvcVolumeMode" json:"persistentStoragePvcVolumeMode" default:"FileSystem"`
	EfsFileSytemIdAndPath          string                  `yaml:"efsFileSytemIdAndPath" json:"efsFileSytemIdAndPath" default:""`
	Secrets                        []string                `yaml:"secrets" json:"secrets" default:"[]"`
	StorageLimit                   string                  `yaml:"storageLimit" json:"storageLimit" default:"5Gi"`
	StorageClass                   string                  `yaml:"storageClass" json:"storageClass" default:"standard"`
	DryRun                         bool                    `yaml:"dryRun" json:"dryRun" default:"false"`
	DnsConfig                      DnsConfig               `yaml:"dns" json:"dns"`
	Resources                      ResourcesConfig         `yaml:"resources" json:"resources"`
	Probes                         ProbesConfig            `yaml:"probes" json:"probes"`
	ServiceMesh                    bool                    `yaml:"serviceMesh" json:"serviceMesh" default:"true"`
	Tls                            bool                    `yaml:"tls" json:"tls" default:"true"`
	DisableTlsLog                  bool                    `yaml:"disableTlsLog" json:"disableTlsLog" default:"true"`
	PacketCapture                  string                  `yaml:"packetCapture" json:"packetCapture" default:"best"`
	Labels                         map[string]string       `yaml:"labels" json:"labels" default:"{}"`
	Annotations                    map[string]string       `yaml:"annotations" json:"annotations" default:"{}"`
	NodeSelectorTerms              NodeSelectorTermsConfig `yaml:"nodeSelectorTerms" json:"nodeSelectorTerms" default:"{}"`
	Tolerations                    TolerationsConfig       `yaml:"tolerations" json:"tolerations" default:"{}"`
	Auth                           AuthConfig              `yaml:"auth" json:"auth"`
	Ingress                        IngressConfig           `yaml:"ingress" json:"ingress"`
	PriorityClass                  string                  `yaml:"priorityClass" json:"priorityClass" default:""`
	Routing                        RoutingConfig           `yaml:"routing" json:"routing"`
	IPv6                           bool                    `yaml:"ipv6" json:"ipv6" default:"true"`
	Debug                          bool                    `yaml:"debug" json:"debug" default:"false"`
	Dashboard                      DashboardConfig         `yaml:"dashboard" json:"dashboard"`
	Telemetry                      TelemetryConfig         `yaml:"telemetry" json:"telemetry"`
	ResourceGuard                  ResourceGuardConfig     `yaml:"resourceGuard" json:"resourceGuard"`
	Watchdog                       WatchdogConfig          `yaml:"watchdog" json:"watchdog"`
	Gitops                         GitopsConfig            `yaml:"gitops" json:"gitops"`
	Sentry                         SentryConfig            `yaml:"sentry" json:"sentry"`
	DefaultFilter                  string                  `yaml:"defaultFilter" json:"defaultFilter" default:"!dns and !error"`
	LiveConfigMapChangesDisabled   bool                    `yaml:"liveConfigMapChangesDisabled" json:"liveConfigMapChangesDisabled" default:"false"`
	GlobalFilter                   string                  `yaml:"globalFilter" json:"globalFilter" default:""`
	EnabledDissectors              []string                `yaml:"enabledDissectors" json:"enabledDissectors"`
	PortMapping                    PortMapping             `yaml:"portMapping" json:"portMapping"`
	CustomMacros                   map[string]string       `yaml:"customMacros" json:"customMacros" default:"{\"https\":\"tls and (http or http2)\"}"`
	Metrics                        MetricsConfig           `yaml:"metrics" json:"metrics"`
	Pprof                          PprofConfig             `yaml:"pprof" json:"pprof"`
	Misc                           MiscConfig              `yaml:"misc" json:"misc"`
	SecurityContext                SecurityContextConfig   `yaml:"securityContext" json:"securityContext"`
	MountBpf                       bool                    `yaml:"mountBpf" json:"mountBpf" default:"true"`
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
