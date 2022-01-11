package shared

const (
	MizuFilteringOptionsEnvVar       = "SENSITIVE_DATA_FILTERING_OPTIONS"
	SyncEntriesConfigEnvVar          = "SYNC_ENTRIES_CONFIG"
	HostModeEnvVar                   = "HOST_MODE"
	NodeNameEnvVar                   = "NODE_NAME"
	TappedAddressesPerNodeDictEnvVar = "TAPPED_ADDRESSES_PER_HOST"
	ConfigDirPath                    = "/app/config/"
	DataDirPath                      = "/app/data/"
	ValidationRulesFileName          = "validation-rules.yaml"
	ContractFileName                 = "contract-oas.yaml"
	ConfigFileName                   = "mizu-config.json"
	GoGCEnvVar                       = "GOGC"
	DefaultApiServerPort             = 8899
	LogLevelEnvVar                   = "LOG_LEVEL"
	BasenineHost                     = "127.0.0.1"
	BaseninePort                     = "9099"
	BasenineImageRepo                = "ghcr.io/up9inc/basenine"
	BasenineImageTag                 = "v0.3.0"
)
