package config

import (
	"errors"
	"fmt"
	"github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/logger"
	"github.com/up9inc/mizu/cli/mizu"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"

	"github.com/creasty/defaults"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/up9inc/mizu/cli/uiUtils"
	"gopkg.in/yaml.v3"
)

const (
	Separator      = "="
	SetCommandName = "set"
	FieldNameTag   = "yaml"
	ReadonlyTag    = "readonly"
)

var (
	Config = ConfigStruct{}
	cmdName string
)

func (config *ConfigStruct) Validate() error {
	if config.IsNsRestrictedMode() {
		if config.Tap.AllNamespaces || len(config.Tap.Namespaces) != 1 || config.Tap.Namespaces[0] != config.MizuResourcesNamespace {
			return fmt.Errorf("Not supported mode. Mizu can't resolve IPs in other namespaces when running in namespace restricted mode.\n"+
				"You can use the same namespace for --%s and --%s", configStructs.NamespacesTapName, MizuResourcesNamespaceConfigName)
		}
	}

	return nil
}

func InitConfig(cmd *cobra.Command) error {
	cmdName = cmd.Name()

	if err := defaults.Set(&Config); err != nil {
		return err
	}

	if err := mergeConfigFile(); err != nil {
		return fmt.Errorf("invalid config %w\n"+
			"you can regenerate the file using `mizu config -r` or just remove it %v", err, GetConfigFilePath())
	}

	cmd.Flags().Visit(initFlag)

	finalConfigPrettified, _ := uiUtils.PrettyJson(Config)
	logger.Log.Debugf("Init config finished\n Final config: %v", finalConfigPrettified)

	return nil
}

func GetConfigWithDefaults() (string, error) {
	defaultConf := ConfigStruct{}
	if err := defaults.Set(&defaultConf); err != nil {
		return "", err
	}

	configElem := reflect.ValueOf(&defaultConf).Elem()
	setZeroForReadonlyFields(configElem)

	return uiUtils.PrettyYaml(defaultConf)
}

func GetConfigFilePath() string {
	return path.Join(mizu.GetMizuFolderPath(), "config.yaml")
}

func mergeConfigFile() error {
	reader, openErr := os.Open(GetConfigFilePath())
	if openErr != nil {
		return nil
	}

	buf, readErr := ioutil.ReadAll(reader)
	if readErr != nil {
		return readErr
	}

	if err := yaml.Unmarshal(buf, &Config); err != nil {
		return err
	}
	logger.Log.Debugf("Found config file, merged to default options")

	return nil
}

func initFlag(f *pflag.Flag) {
	configElem := reflect.ValueOf(&Config).Elem()

	sliceValue, isSliceValue := f.Value.(pflag.SliceValue)
	if !isSliceValue {
		flagPath := getFlagPath(f.Name)
		if err := mergeFlagValue(configElem, flagPath, strings.Join(flagPath, "."), f.Value.String()); err != nil {
			logger.Log.Warningf(uiUtils.Warning, err)
		}
		return
	}

	if f.Name == SetCommandName {
		mergeSetFlag(configElem, sliceValue.GetSlice())
		return
	}

	flagPath := getFlagPath(f.Name)
	if err := mergeFlagValues(configElem, flagPath, strings.Join(flagPath, "."), sliceValue.GetSlice()); err != nil {
		logger.Log.Warningf(uiUtils.Warning, err)
	}
}

func getFlagPath(flagName string) []string {
	flagPath := []string {cmdName}
	return append(flagPath, strings.Split(flagName, ".")...)
}

func mergeSetFlag(configElem reflect.Value, setValues []string) {
	setMap := map[string][]string{}

	for _, setValue := range setValues {
		if !strings.Contains(setValue, Separator) {
			logger.Log.Warningf(uiUtils.Warning, fmt.Sprintf("Ignoring set argument %s (set argument format: <flag name>=<flag value>)", setValue))
			continue
		}

		split := strings.SplitN(setValue, Separator, 2)
		if len(split) != 2 {
			logger.Log.Warningf(uiUtils.Warning, fmt.Sprintf("Ignoring set argument %s (set argument format: <flag name>=<flag value>)", setValue))
			continue
		}

		argumentKey, argumentValue := split[0], split[1]

		setMap[argumentKey] = append(setMap[argumentKey], argumentValue)
	}

	for argumentKey, argumentValues := range setMap {
		flagPath := strings.Split(argumentKey, ".")

		if len(argumentValues) > 1 {
			if err := mergeFlagValues(configElem, flagPath, argumentKey, argumentValues); err != nil {
				logger.Log.Warningf(uiUtils.Warning, err)
			}
		} else {
			if err := mergeFlagValue(configElem, flagPath, argumentKey, argumentValues[0]); err != nil {
				logger.Log.Warningf(uiUtils.Warning, err)
			}
		}
	}
}

func mergeFlagValue(currentElem reflect.Value, flagPath []string, fullFlagName string, flagValue string) error {
	mergeFunction := func(flagName string, currentField reflect.StructField, currentFieldByName reflect.Value, currentElem reflect.Value) error {
		currentFieldKind := currentField.Type.Kind()

		if currentFieldKind == reflect.Slice {
			return mergeFlagValues(currentElem, []string{flagName}, fullFlagName, []string{flagValue})
		}

		parsedValue, err := getParsedValue(currentFieldKind, flagValue)
		if err != nil {
			return fmt.Errorf("invalid value %s for flag name %s, expected %s", flagValue, flagName, currentFieldKind)
		}

		currentFieldByName.Set(parsedValue)
		return nil
	}

	return mergeFlag(currentElem, flagPath, fullFlagName, mergeFunction)
}

func mergeFlagValues(currentElem reflect.Value, flagPath []string, fullFlagName string, flagValues []string) error {
	mergeFunction := func(flagName string, currentField reflect.StructField, currentFieldByName reflect.Value, currentElem reflect.Value) error {
		currentFieldKind := currentField.Type.Kind()

		if currentFieldKind != reflect.Slice {
			return fmt.Errorf("invalid values %s for flag name %s, expected %s", strings.Join(flagValues, ","), flagName, currentFieldKind)
		}

		flagValueKind := currentField.Type.Elem().Kind()

		parsedValues := reflect.MakeSlice(reflect.SliceOf(currentField.Type.Elem()), 0, 0)
		for _, flagValue := range flagValues {
			parsedValue, err := getParsedValue(flagValueKind, flagValue)
			if err != nil {
				return fmt.Errorf("invalid value %s for flag name %s, expected %s", flagValue, flagName, flagValueKind)
			}

			parsedValues = reflect.Append(parsedValues, parsedValue)
		}

		currentFieldByName.Set(parsedValues)
		return nil
	}

	return mergeFlag(currentElem, flagPath, fullFlagName, mergeFunction)
}

func mergeFlag(currentElem reflect.Value, currentFlagPath []string, fullFlagName string, mergeFunction func(flagName string, currentField reflect.StructField, currentFieldByName reflect.Value, currentElem reflect.Value) error) error {
	if len(currentFlagPath) == 0 {
		return fmt.Errorf("flag \"%s\" not found", fullFlagName)
	}

	for i := 0; i < currentElem.NumField(); i++ {
		currentField := currentElem.Type().Field(i)
		currentFieldByName := currentElem.FieldByName(currentField.Name)

		if currentField.Type.Kind() == reflect.Struct && getFieldNameByTag(currentField) == currentFlagPath[0] {
			return mergeFlag(currentFieldByName, currentFlagPath[1:], fullFlagName, mergeFunction)
		}

		if len(currentFlagPath) > 1 || getFieldNameByTag(currentField) != currentFlagPath[0] {
			continue
		}

		return mergeFunction(currentFlagPath[0], currentField, currentFieldByName, currentElem)
	}

	return fmt.Errorf("flag \"%s\" not found", fullFlagName)
}

func getFieldNameByTag(field reflect.StructField) string {
	return strings.Split(field.Tag.Get(FieldNameTag), ",")[0]
}

func getParsedValue(kind reflect.Kind, value string) (reflect.Value, error) {
	switch kind {
	case reflect.String:
		return reflect.ValueOf(value), nil
	case reflect.Bool:
		boolArgumentValue, err := strconv.ParseBool(value)
		if err != nil {
			break
		}

		return reflect.ValueOf(boolArgumentValue), nil
	case reflect.Int:
		intArgumentValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			break
		}

		return reflect.ValueOf(int(intArgumentValue)), nil
	case reflect.Int8:
		intArgumentValue, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			break
		}

		return reflect.ValueOf(int8(intArgumentValue)), nil
	case reflect.Int16:
		intArgumentValue, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			break
		}

		return reflect.ValueOf(int16(intArgumentValue)), nil
	case reflect.Int32:
		intArgumentValue, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			break
		}

		return reflect.ValueOf(int32(intArgumentValue)), nil
	case reflect.Int64:
		intArgumentValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			break
		}

		return reflect.ValueOf(intArgumentValue), nil
	case reflect.Uint:
		uintArgumentValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			break
		}

		return reflect.ValueOf(uint(uintArgumentValue)), nil
	case reflect.Uint8:
		uintArgumentValue, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			break
		}

		return reflect.ValueOf(uint8(uintArgumentValue)), nil
	case reflect.Uint16:
		uintArgumentValue, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			break
		}

		return reflect.ValueOf(uint16(uintArgumentValue)), nil
	case reflect.Uint32:
		uintArgumentValue, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			break
		}

		return reflect.ValueOf(uint32(uintArgumentValue)), nil
	case reflect.Uint64:
		uintArgumentValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			break
		}

		return reflect.ValueOf(uintArgumentValue), nil
	}

	return reflect.ValueOf(nil), errors.New("value to parse does not match type")
}

func setZeroForReadonlyFields(currentElem reflect.Value) {
	for i := 0; i < currentElem.NumField(); i++ {
		currentField := currentElem.Type().Field(i)
		currentFieldByName := currentElem.FieldByName(currentField.Name)

		if currentField.Type.Kind() == reflect.Struct {
			setZeroForReadonlyFields(currentFieldByName)
			continue
		}

		if _, ok := currentField.Tag.Lookup(ReadonlyTag); ok {
			currentFieldByName.Set(reflect.Zero(currentField.Type))
		}
	}
}
