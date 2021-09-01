package config

import (
	"fmt"
	"github.com/up9inc/mizu/cli/logger"
	"github.com/up9inc/mizu/cli/uiUtils"
	"os"
	"reflect"
)

const (
	ApiServerRetries = "API_SERVER_RETRIES"
)

func GetEnvConfig(kind reflect.Kind, key string, defaultValue reflect.Value) reflect.Value {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}

	parsedValue, err := getParsedValue(kind, val)
	if err != nil {
		logger.Log.Warningf(uiUtils.Warning, fmt.Sprintf("error parsing env value, expected value: %v, key: %v, value: %v", kind, key, val))
		return defaultValue
	}

	return parsedValue
}
