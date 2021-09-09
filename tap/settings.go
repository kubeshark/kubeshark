package tap

import (
	"os"
	"strconv"
)

const (
	MemoryProfilingEnabledEnvVarName          = "MEMORY_PROFILING_ENABLED"
	MemoryProfilingDumpPath                   = "MEMORY_PROFILING_DUMP_PATH"
	MemoryProfilingTimeIntervalSeconds        = "MEMORY_PROFILING_TIME_INTERVAL"
	MaxBufferedPagesTotalEnvVarName           = "MAX_BUFFERED_PAGES_TOTAL"
	MaxBufferedPagesPerConnectionEnvVarName   = "MAX_BUFFERED_PAGES_PER_CONNECTION"
	MaxBufferedPagesTotalDefaultValue         = 5000
	MaxBufferedPagesPerConnectionDefaultValue = 5000
)

type globalSettings struct {
	filterAuthorities []string
}

var gSettings = &globalSettings{
	filterAuthorities: []string{},
}

func SetFilterAuthorities(ipAddresses []string) {
	gSettings.filterAuthorities = ipAddresses
}

func GetFilterIPs() []string {
	addresses := make([]string, len(gSettings.filterAuthorities))
	copy(addresses, gSettings.filterAuthorities)
	return addresses
}

func GetMaxBufferedPagesTotal() int {
	valueFromEnv, err := strconv.Atoi(os.Getenv(MaxBufferedPagesTotalEnvVarName))
	if err != nil {
		return MaxBufferedPagesTotalDefaultValue
	}
	return valueFromEnv
}

func GetMaxBufferedPagesPerConnection() int {
	valueFromEnv, err := strconv.Atoi(os.Getenv(MaxBufferedPagesPerConnectionEnvVarName))
	if err != nil {
		return MaxBufferedPagesPerConnectionDefaultValue
	}
	return valueFromEnv
}

func GetMemoryProfilingEnabled() bool {
	return os.Getenv(MemoryProfilingEnabledEnvVarName) == "1"
}
