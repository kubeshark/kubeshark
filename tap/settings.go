package tap

import (
	"os"
	"strconv"
	"time"

	"github.com/up9inc/mizu/logger"
)

const (
	MemoryProfilingEnabledEnvVarName               = "MEMORY_PROFILING_ENABLED"
	MemoryProfilingDumpPath                        = "MEMORY_PROFILING_DUMP_PATH"
	MemoryProfilingTimeIntervalSeconds             = "MEMORY_PROFILING_TIME_INTERVAL"
	MaxBufferedPagesTotalEnvVarName                = "MAX_BUFFERED_PAGES_TOTAL"
	MaxBufferedPagesPerConnectionEnvVarName        = "MAX_BUFFERED_PAGES_PER_CONNECTION"
	MaxBufferedPagesTotalDefaultValue              = 5000
	MaxBufferedPagesPerConnectionDefaultValue      = 5000
	TcpStreamChannelTimeoutMsEnvVarName            = "TCP_STREAM_CHANNEL_TIMEOUT_MS"
	TcpStreamChannelTimeoutMsDefaultValue          = 10000
	CloseTimedoutTcpChannelsIntervalMsEnvVarName   = "CLOSE_TIMEDOUT_TCP_STREAM_CHANNELS_INTERVAL_MS"
	CloseTimedoutTcpChannelsIntervalMsDefaultValue = 1000
	CloseTimedoutTcpChannelsIntervalMsMinValue     = 10
	CloseTimedoutTcpChannelsIntervalMsMaxValue     = 10000
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

func GetMemoryProfilingEnabled() bool {
	return os.Getenv(MemoryProfilingEnabledEnvVarName) == "1"
}

func GetTcpChannelTimeoutMs() time.Duration {
	valueFromEnv, err := strconv.Atoi(os.Getenv(TcpStreamChannelTimeoutMsEnvVarName))
	if err != nil {
		return TcpStreamChannelTimeoutMsDefaultValue * time.Millisecond
	}
	return time.Duration(valueFromEnv) * time.Millisecond
}

func GetCloseTimedoutTcpChannelsInterval() time.Duration {
	defaultDuration := CloseTimedoutTcpChannelsIntervalMsDefaultValue * time.Millisecond
	rangeMin := CloseTimedoutTcpChannelsIntervalMsMinValue
	rangeMax := CloseTimedoutTcpChannelsIntervalMsMaxValue
	closeTimedoutTcpChannelsIntervalMsStr := os.Getenv(CloseTimedoutTcpChannelsIntervalMsEnvVarName)
	if closeTimedoutTcpChannelsIntervalMsStr == "" {
		return defaultDuration
	} else {
		closeTimedoutTcpChannelsIntervalMs, err := strconv.Atoi(closeTimedoutTcpChannelsIntervalMsStr)
		if err != nil {
			logger.Log.Warningf("Error parsing environment variable %s: %v\n", CloseTimedoutTcpChannelsIntervalMsEnvVarName, err)
			return defaultDuration
		} else {
			if closeTimedoutTcpChannelsIntervalMs < rangeMin || closeTimedoutTcpChannelsIntervalMs > rangeMax {
				logger.Log.Warningf("The value of environment variable %s is not in acceptable range: %d - %d\n", CloseTimedoutTcpChannelsIntervalMsEnvVarName, rangeMin, rangeMax)
				return defaultDuration
			} else {
				return time.Duration(closeTimedoutTcpChannelsIntervalMs) * time.Millisecond
			}
		}
	}
}
