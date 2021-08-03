package tap

import (
	"os"
	"strconv"
)

const (
	MaxBufferedPagesTotal                     = "MAX_BUFFERED_PAGES_TOTAL"
	MaxBufferedPagesPerConnection             = "MAX_BUFFERED_PAGES_PER_CONNECTION"
	MaxBufferedPagesTotalDefaultValue         = "100000"
	MaxBufferedPagesPerConnectionDefaultValue = "100000"
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
	valueFromEnv, err := strconv.Atoi(getFromEnvWithDefault(MaxBufferedPagesTotal, MaxBufferedPagesTotalDefaultValue))
	if err != nil {
		return 0
	}
	return valueFromEnv
}

func GetMaxBufferedPagesPerConnection() int {
	valueFromEnv, err := strconv.Atoi(getFromEnvWithDefault(MaxBufferedPagesPerConnection, MaxBufferedPagesPerConnectionDefaultValue))
	if err != nil {
		return 0
	}
	return valueFromEnv
}

func GetMemoryProfilingEnabled() bool {
	return os.Getenv("MEMORY_PROFILING_ENABLED") == "1"
}

func getFromEnvWithDefault(key string, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

