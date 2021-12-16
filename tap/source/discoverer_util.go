package source

import (
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/up9inc/mizu/shared/logger"
)

var numberRegex = regexp.MustCompile("[0-9]+")

func readEnvironmentVariable(file string, name string) (string, error) {
	bytes, err := ioutil.ReadFile(file)

	if err != nil {
		logger.Log.Warningf("Error reading environment file %v - %v", file, err)
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

		if name == varName {
			return value, nil
		}
	}

	return "", nil
}
