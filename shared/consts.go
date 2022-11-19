package shared

const (
	KubesharkFilteringOptionsEnvVar = "SENSITIVE_DATA_FILTERING_OPTIONS"
	HostModeEnvVar                  = "HOST_MODE"
	NodeNameEnvVar                  = "NODE_NAME"
	ConfigDirPath                   = "/app/config/"
	DataDirPath                     = "/app/data/"
	ConfigFileName                  = "kubeshark-config.json"
	DefaultApiServerPort            = 8899
	LogLevelEnvVar                  = "LOG_LEVEL"
	KubesharkAgentImageRepo         = "docker.io/kubeshark/kubeshark"
	BasenineHost                    = "127.0.0.1"
	BaseninePort                    = "9099"
	BasenineReconnectInterval       = 3
)
