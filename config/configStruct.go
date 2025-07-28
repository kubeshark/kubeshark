package config

import (
	"os"
	"path/filepath"

	"github.com/kubeshark/kubeshark/config/configStructs"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/homedir"
)

const (
	KubeConfigPathConfigName = "kube-configPath"
)

func CreateDefaultConfig() ConfigStruct {
	return ConfigStruct{
		Tap: configStructs.TapConfig{
			NodeSelectorTerms: configStructs.NodeSelectorTermsConfig{
				Workers: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{
							{
								Key:      "kubernetes.io/os",
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{"linux"},
							},
						},
					},
				},
				Hub: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{
							{
								Key:      "kubernetes.io/os",
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{"linux"},
							},
						},
					},
				},
				Front: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{
							{
								Key:      "kubernetes.io/os",
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{"linux"},
							},
						},
					},
				},
				Dex: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{
							{
								Key:      "kubernetes.io/os",
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{"linux"},
							},
						},
					},
				},
			},
			Tolerations: configStructs.TolerationsConfig{
				Workers: []v1.Toleration{
					{
						Effect:   v1.TaintEffect("NoExecute"),
						Operator: v1.TolerationOpExists,
					},
				},
			},
			SecurityContext: configStructs.SecurityContextConfig{
				Privileged: true,
				// Capabilities used only when running in unprivileged mode
				Capabilities: configStructs.CapabilitiesConfig{
					NetworkCapture: []string{
						// NET_RAW is required to listen the network traffic
						"NET_RAW",
						// NET_ADMIN is required to listen the network traffic
						"NET_ADMIN",
					},
					ServiceMeshCapture: []string{
						// SYS_ADMIN is required to read /proc/PID/net/ns + to install eBPF programs (kernel < 5.8)
						"SYS_ADMIN",
						// SYS_PTRACE is required to set netns to other process + to open libssl.so of other process
						"SYS_PTRACE",
						// DAC_OVERRIDE is required to read /proc/PID/environ
						"DAC_OVERRIDE",
					},
					EBPFCapture: []string{
						// SYS_ADMIN is required to read /proc/PID/net/ns + to install eBPF programs (kernel < 5.8)
						"SYS_ADMIN",
						// SYS_PTRACE is required to set netns to other process + to open libssl.so of other process
						"SYS_PTRACE",
						// SYS_RESOURCE is required to change rlimits for eBPF
						"SYS_RESOURCE",
						// IPC_LOCK is required for ebpf perf buffers allocations after some amount of size buffer size:
						// https://github.com/kubeshark/tracer/blob/13e24725ba8b98216dd0e553262e6d9c56dce5fa/main.go#L82)
						"IPC_LOCK",
					},
				},
			},
			Auth: configStructs.AuthConfig{
				Saml: configStructs.SamlConfig{
					RoleAttribute: "role",
					Roles: map[string]configStructs.Role{
						"admin": {
							Filter:          "",
							CanDownloadPCAP: true,
							CanUseScripting: true,
							ScriptingPermissions: configStructs.ScriptingPermissions{
								CanSave:     true,
								CanActivate: true,
								CanDelete:   true,
							},
							CanUpdateTargetedPods:   true,
							CanStopTrafficCapturing: true,
							ShowAdminConsoleLink:    true,
						},
					},
				},
			},
			EnabledDissectors: []string{
				"amqp",
				"dns",
				"http",
				"icmp",
				"kafka",
				"redis",
				// "sctp",
				// "syscall",
				// "tcp",
				// "udp",
				"ws",
				// "tlsx",
				"ldap",
				"radius",
				"diameter",
			},
			PortMapping: configStructs.PortMapping{
				HTTP:     []uint16{80, 443, 8080},
				AMQP:     []uint16{5671, 5672},
				KAFKA:    []uint16{9092},
				REDIS:    []uint16{6379},
				LDAP:     []uint16{389},
				DIAMETER: []uint16{3868},
			},
			Dashboard: configStructs.DashboardConfig{
				CompleteStreamingEnabled: true,
			},
		},
	}
}

type KubeConfig struct {
	ConfigPathStr string `yaml:"configPath" json:"configPath"`
	Context       string `yaml:"context" json:"context"`
}

type ManifestsConfig struct {
	Dump bool `yaml:"dump" json:"dump"`
}

type ConfigStruct struct {
	Tap                  configStructs.TapConfig       `yaml:"tap" json:"tap"`
	Logs                 configStructs.LogsConfig      `yaml:"logs" json:"logs"`
	Config               configStructs.ConfigConfig    `yaml:"config,omitempty" json:"config,omitempty"`
	PcapDump             configStructs.PcapDumpConfig  `yaml:"pcapdump" json:"pcapdump"`
	Kube                 KubeConfig                    `yaml:"kube" json:"kube"`
	DumpLogs             bool                          `yaml:"dumpLogs" json:"dumpLogs" default:"false"`
	HeadlessMode         bool                          `yaml:"headless" json:"headless" default:"false"`
	License              string                        `yaml:"license" json:"license" default:""`
	CloudLicenseEnabled  bool                          `yaml:"cloudLicenseEnabled" json:"cloudLicenseEnabled" default:"true"`
	AiAssistantEnabled   bool                          `yaml:"aiAssistantEnabled" json:"aiAssistantEnabled" default:"true"`
	DemoModeEnabled      bool                          `yaml:"demoModeEnabled" json:"demoModeEnabled" default:"false"`
	SupportChatEnabled   bool                          `yaml:"supportChatEnabled" json:"supportChatEnabled" default:"true"`
	BetaEnabled          bool                          `yaml:"betaEnabled" json:"betaEnabled" default:"false"`
	InternetConnectivity bool                          `yaml:"internetConnectivity" json:"internetConnectivity" default:"true"`
	Scripting            configStructs.ScriptingConfig `yaml:"scripting" json:"scripting"`
	Manifests            ManifestsConfig               `yaml:"manifests,omitempty" json:"manifests,omitempty"`
	Timezone             string                        `yaml:"timezone" json:"timezone"`
	LogLevel             string                        `yaml:"logLevel" json:"logLevel" default:"warning"`
}

func (config *ConfigStruct) ImagePullPolicy() v1.PullPolicy {
	return v1.PullPolicy(config.Tap.Docker.ImagePullPolicy)
}

func (config *ConfigStruct) ImagePullSecrets() []v1.LocalObjectReference {
	var ref []v1.LocalObjectReference
	for _, name := range config.Tap.Docker.ImagePullSecrets {
		ref = append(ref, v1.LocalObjectReference{Name: name})
	}

	return ref
}

func (config *ConfigStruct) KubeConfigPath() string {
	if config.Kube.ConfigPathStr != "" {
		return config.Kube.ConfigPathStr
	}

	envKubeConfigPath := os.Getenv("KUBECONFIG")
	if envKubeConfigPath != "" {
		return envKubeConfigPath
	}

	home := homedir.HomeDir()
	return filepath.Join(home, ".kube", "config")
}
