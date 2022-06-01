package shared

const (
	MizuFilteringOptionsEnvVar = "SENSITIVE_DATA_FILTERING_OPTIONS"
	HostModeEnvVar             = "HOST_MODE"
	NodeNameEnvVar             = "NODE_NAME"
	ConfigDirPath              = "/app/config/"
	DataDirPath                = "/app/data/"
	ValidationRulesFileName    = "validation-rules.yaml"
	ContractFileName           = "contract-oas.yaml"
	ConfigFileName             = "mizu-config.json"
	DefaultApiServerPort       = 8899
	LogLevelEnvVar             = "LOG_LEVEL"
	MizuAgentImageRepo         = "docker.io/up9inc/mizu"
	BasenineHost               = "127.0.0.1"
	BaseninePort               = "9099"
	BasenineReconnectInterval  = 3
)
