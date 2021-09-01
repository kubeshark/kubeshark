package config

import (
	"os"
	"strconv"
)

const (
	ApiServerRetries = "API_SERVER_RETRIES"
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
