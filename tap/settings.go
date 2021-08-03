package tap

import (
	"os"
	"strconv"
)

const (
	MemoryProfilingEnabledEnvVarName          = "MEMORY_PROFILING_ENABLED"
	MaxBufferedPagesTotalEnvVarName           = "MAX_BUFFERED_PAGES_TOTAL"
	MaxBufferedPagesPerConnectionEnvVarName   = "MAX_BUFFERED_PAGES_PER_CONNECTION"
	MaxBufferedPagesTotalDefaultValue         = 5000
	MaxBufferedPagesPerConnectionDefaultValue = 5000
)

type globalSettings struct {
	filterPorts       []int
	filterAuthorities []string
}

var gSettings = &globalSettings{
	filterPorts:       []int{},
	filterAuthorities: []string{},
}

func SetFilterPorts(ports []int) {
	gSettings.filterPorts = ports
}

func GetFilterPorts() []int {
	ports := make([]int, len(gSettings.filterPorts))
	copy(ports, gSettings.filterPorts)
	return ports
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
