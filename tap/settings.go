package tap

import (
	"os"
	"strconv"
	"time"
)

const (
	MemoryProfilingEnabledEnvVarName          = "MEMORY_PROFILING_ENABLED"
	MemoryProfilingDumpPath                   = "MEMORY_PROFILING_DUMP_PATH"
	MemoryProfilingTimeIntervalSeconds        = "MEMORY_PROFILING_TIME_INTERVAL"
	MaxBufferedPagesTotalEnvVarName           = "MAX_BUFFERED_PAGES_TOTAL"
	MaxBufferedPagesPerConnectionEnvVarName   = "MAX_BUFFERED_PAGES_PER_CONNECTION"
	TcpStreamChannelTimeoutMsEnvVarName       = "TCP_STREAM_CHANNEL_TIMEOUT_MS"
	MaxBufferedPagesTotalDefaultValue         = 5000
	MaxBufferedPagesPerConnectionDefaultValue = 5000
	TcpStreamChannelTimeoutMsDefaultValue     = 10000
)

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

func GetTcpChannelTimeoutMs() time.Duration {
	valueFromEnv, err := strconv.Atoi(os.Getenv(TcpStreamChannelTimeoutMsEnvVarName))
	if err != nil {
		return TcpStreamChannelTimeoutMsDefaultValue * time.Millisecond
	}
	return time.Duration(valueFromEnv) * time.Millisecond
}

func GetMemoryProfilingEnabled() bool {
	return os.Getenv(MemoryProfilingEnabledEnvVarName) == "1"
}
