package mizu

import (
	"errors"
	"fmt"
	"github.com/up9inc/mizu/cli/uiUtils"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
)

const separator = "="

var configObj = map[string]interface{}{}

type CommandLineFlag struct {
	CommandLineName   string
	YamlHierarchyName string
	DefaultValue      interface{}
	Type              reflect.Kind
}

const (
	ConfigurationKeyAnalyzingDestination = "tap.dest"
	ConfigurationKeyUploadInterval       = "tap.uploadInterval"
	ConfigurationKeyMizuImage            = "mizuImage"
)

var allowedSetFlags = []CommandLineFlag{
	{
		CommandLineName:   "dest",
		YamlHierarchyName: ConfigurationKeyAnalyzingDestination,
		DefaultValue:      "up9.app",
		Type:              reflect.String,
		// TODO: maybe add short description that we can show
	},
	{
		CommandLineName:   "uploadInterval",
		YamlHierarchyName: ConfigurationKeyUploadInterval,
		DefaultValue:      10,
		Type:              reflect.Int,
	},
	{
		CommandLineName:   "mizuImage",
		YamlHierarchyName: ConfigurationKeyMizuImage,
		DefaultValue:      fmt.Sprintf("gcr.io/up9-docker-hub/mizu/%s:%s", Branch, SemVer),
		Type:              reflect.String,
	},
}

func GetString(key string) string {
	return fmt.Sprintf("%v", getValueFromMergedConfig(key))
}

func GetInt(key string) int {
	stringVal := GetString(key)
	Log.Debugf("Found string value %v", stringVal)

	val, err := strconv.Atoi(stringVal)
	if err != nil {
		Log.Warningf("Invalid value %v for key %s", val, key)
		os.Exit(1)
	}
	return val
}

func InitConfig(commandLineValues []string) error {
	Log.Debugf("Merging default values")
	mergeDefaultValues()
	Log.Debugf("Merging config file values")
	if err1 := mergeConfigFile(); err1 != nil {
		Log.Infof(fmt.Sprintf(uiUtils.Red, "Invalid config file\n"))
		return err1
	}
	Log.Debugf("Merging command line values")
	if err2 := mergeCommandLineFlags(commandLineValues); err2 != nil {
		Log.Infof(fmt.Sprintf(uiUtils.Red, "Invalid commanad argument\n"))
		return err2
	}
	finalConfigPrettified, _ := uiUtils.PrettyJson(configObj)
	Log.Debugf("Merged all config successfully\n Final config: %v", finalConfigPrettified)
	return nil
}

func GetTemplateConfig() string {
	templateConfig := map[string]interface{}{}
	for _, allowedFlag := range allowedSetFlags {
		addToConfigObj(allowedFlag.YamlHierarchyName, allowedFlag.DefaultValue, templateConfig)
	}
	prettifiedConfig, _ := uiUtils.PrettyYaml(templateConfig)
	return prettifiedConfig
}

func GetConfigStr() string {
	val, _ := uiUtils.PrettyYaml(configObj)
	return val
}

func getValueFromMergedConfig(key string) interface{} {
	if a, ok := configObj[key]; ok {
		return a
	}
	return nil
}

func mergeDefaultValues() {
	for _, allowedFlag := range allowedSetFlags {
		Log.Debugf("Setting %v to %v", allowedFlag.YamlHierarchyName, allowedFlag.DefaultValue)
		configObj[allowedFlag.YamlHierarchyName] = allowedFlag.DefaultValue
	}
}

func mergeConfigFile() error {
	Log.Debugf("Merging mizu config file values")
	home, homeDirErr := os.UserHomeDir()
	if homeDirErr != nil {
		return nil
	}
	reader, openErr := os.Open(path.Join(home, ".mizu", "config.yaml"))
	if openErr != nil {
		return nil
	}
	buf, readErr := ioutil.ReadAll(reader)
	if readErr != nil {
		return readErr
	}
	m := make(map[string]interface{})
	if err := yaml.Unmarshal(buf, &m); err != nil {
		return err
	}
	for k, v := range m {
		addToConfig(k, v)
	}
	return nil
}

func addToConfig(prefix string, value interface{}) {
	typ := reflect.TypeOf(value).Kind()
	if typ == reflect.Int || typ == reflect.String || typ == reflect.Slice {
		validateConfigFileKey(prefix)
		configObj[prefix] = value
	} else if typ == reflect.Map {
		for k1, v1 := range value.(map[string]interface{}) {
			addToConfig(fmt.Sprintf("%s.%s", prefix, k1), v1)
		}
	}
}

func mergeCommandLineFlags(commandLineValues []string) error {
	Log.Debugf("Merging Command line flags")
	for _, e := range commandLineValues {
		if !strings.Contains(e, separator) {
			return errors.New(fmt.Sprintf("invalid set argument %s", e))
		}
		split := strings.SplitN(e, separator, 2)
		if len(split) != 2 {
			return errors.New(fmt.Sprintf("invalid set argument %s", e))
		}
		setFlagKey, argumentValue := split[0], split[1]
		argumentNameInConfig, expectedType, err := flagFromAllowed(setFlagKey)
		if err != nil {
			return err
		}
		argumentType := reflect.ValueOf(argumentValue).Kind()
		if argumentType != expectedType {
			return errors.New(fmt.Sprintf("Invalid value for argument %s (should be type %s but got %s", setFlagKey, expectedType, argumentType))
		}
		configObj[argumentNameInConfig] = argumentValue
	}
	return nil
}

func flagFromAllowed(setFlagKey string) (string, reflect.Kind, error) {
	for _, allowedFlag := range allowedSetFlags {
		if strings.ToLower(allowedFlag.CommandLineName) == strings.ToLower(setFlagKey) {
			return allowedFlag.YamlHierarchyName, allowedFlag.Type, nil
		}
	}
	return "", reflect.Invalid, errors.New(fmt.Sprintf("invalid set argument %s", setFlagKey))
}

func validateConfigFileKey(configFileKey string) {
	for _, allowedFlag := range allowedSetFlags {
		if allowedFlag.YamlHierarchyName == configFileKey {
			return
		}
	}
	Log.Info(fmt.Sprintf("Unknown argument: %s. Exit", configFileKey))
	os.Exit(1)
}

func addToConfigObj(key string, value interface{}, configObj map[string]interface{}) {
	typ := reflect.TypeOf(value).Kind()
	if typ == reflect.Int || typ == reflect.String || typ == reflect.Slice {
		if strings.Contains(key, ".") {
			split := strings.SplitN(key, ".", 2)
			firstLevelKey := split[0]
			if _, ok := configObj[firstLevelKey]; !ok {
				configObj[firstLevelKey] = map[string]interface{}{}
			}
			addToConfigObj(split[1], value, configObj[firstLevelKey].(map[string]interface{}))
		} else {
			configObj[key] = value
		}
	}
}

