package mizu

import (
	"errors"
	"fmt"
	"github.com/creasty/defaults"
	"github.com/spf13/pflag"
	"github.com/up9inc/mizu/cli/mizu/configStructs"
	"github.com/up9inc/mizu/cli/uiUtils"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
)

const (
	Separator = "="
	SetCommandName = "set"
)

var Config = configStructs.ConfigStruct{}

func InitConfig() error {
	if err := defaults.Set(&Config); err != nil {
		return err
	}

	if err := mergeConfigFile(); err != nil {
		Log.Infof(fmt.Sprintf(uiUtils.Red, "Invalid config file"))
		return err
	}

	finalConfigPrettified, _ := uiUtils.PrettyJson(Config)
	Log.Debugf("Merged all config successfully\n Final config: %v", finalConfigPrettified)

	return nil
}

func InitFlag(f *pflag.Flag) {
	configElem := reflect.ValueOf(&Config).Elem()

	sliceValue, isSliceValue := f.Value.(pflag.SliceValue)
	if !isSliceValue {
		mergeFlagValue(configElem, f.Name, f.Value.String())
		return
	}

	if f.Name == SetCommandName {
		if setError := mergeSetFlag(sliceValue.GetSlice()); setError != nil {
			Log.Infof(fmt.Sprintf(uiUtils.Red, "Invalid set argument"))
		}
		return
	}

	for _, value := range sliceValue.GetSlice() {
		mergeFlagValue(configElem, f.Name, value)
	}
}

func GetTemplateConfig() string {
	prettifiedConfig, _ := uiUtils.PrettyYaml(Config)
	return prettifiedConfig
}

func mergeConfigFile() error {
	Log.Debugf("Merging config file values")
	home, homeDirErr := os.UserHomeDir()
	if homeDirErr != nil {
		return homeDirErr
	}

	reader, openErr := os.Open(path.Join(home, ".mizu", "config.yaml"))
	if openErr != nil {
		return openErr
	}

	buf, readErr := ioutil.ReadAll(reader)
	if readErr != nil {
		return readErr
	}

	if err := yaml.Unmarshal(buf, &Config); err != nil {
		return err
	}

	return nil
}

func mergeSetFlag(setValues []string) error {
	configElem := reflect.ValueOf(&Config).Elem()

	for _, setValue := range setValues {
		if !strings.Contains(setValue, Separator) {
			return errors.New(fmt.Sprintf("invalid set argument %s", setValue))
		}

		split := strings.SplitN(setValue, Separator, 2)
		if len(split) != 2 {
			return errors.New(fmt.Sprintf("invalid set argument %s", setValue))
		}

		argumentKey, argumentValue := split[0], split[1]
		mergeFlagValue(configElem, argumentKey, argumentValue)
	}

	return nil
}

func mergeFlagValue(currentElem reflect.Value, flagKey string, flagValue string) {
	for i := 0; i < currentElem.NumField(); i++ {
		currentField := currentElem.Type().Field(i)

		if currentField.Type.Kind() == reflect.Struct {
			mergeFlagValue(currentElem.FieldByName(currentField.Name), flagKey, flagValue)
			continue
		}

		if currentField.Tag.Get("yaml") != flagKey {
			continue
		}

		switch currentField.Type.Kind() {
		case reflect.String:
			currentElem.FieldByName(currentField.Name).SetString(flagValue)
		case reflect.Bool:
			boolArgumentValue, err := strconv.ParseBool(flagValue)
			if err != nil {
				Log.Warningf(uiUtils.Red, fmt.Sprintf("Invalid value %v for key %s, expected bool", flagValue, flagKey))
			} else {
				currentElem.FieldByName(currentField.Name).SetBool(boolArgumentValue)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intArgumentValue, err := strconv.ParseInt(flagValue, 10, 64)
			if err != nil {
				Log.Warningf(uiUtils.Red, fmt.Sprintf("Invalid value %v for key %s, expected int", flagValue, flagKey))
			} else {
				currentElem.FieldByName(currentField.Name).SetInt(intArgumentValue)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uintArgumentValue, err := strconv.ParseUint(flagValue, 10, 64)
			if err != nil {
				Log.Warningf(uiUtils.Red, fmt.Sprintf("Invalid value %v for key %s, expected uint", flagValue, flagKey))
			} else {
				currentElem.FieldByName(currentField.Name).SetUint(uintArgumentValue)
			}
		}
	}
}
