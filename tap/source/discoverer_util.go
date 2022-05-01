package source

import (
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/up9inc/mizu/logger"
)

var numberRegex = regexp.MustCompile("[0-9]+")

func getSingleValueFromEnvironmentVariableFile(filePath string, variableName string) (string, error) {
	bytes, err := ioutil.ReadFile(filePath)

	if err != nil {
		logger.Log.Warningf("Error reading environment file %v - %v", filePath, err)
		return "", err
	}

	envs := strings.Split(string(bytes), string([]byte{0}))

	for _, env := range envs {
		if !strings.Contains(env, "=") {
			continue
		}

		parts := strings.Split(env, "=")
		varName := parts[0]
		value := parts[1]

		if variableName == varName {
			return value, nil
		}
	}

	return "", nil
}
