package config

import (
	"os"
	"strconv"
)

const (
	HubRetries    = "HUB_SERVER_RETRIES"
	HubTimeoutSec = "HUB_SERVER_TIMEOUT_SEC"
)

func GetIntEnvConfig(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}
