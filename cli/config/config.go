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
	configElemValue := reflect.ValueOf(&Config).Elem()

	flagPath := []string {cmdName, f.Name}

	sliceValue, isSliceValue := f.Value.(pflag.SliceValue)
	if !isSliceValue {
		if err := mergeFlagValue(configElemValue, flagPath, strings.Join(flagPath, "."), f.Value.String()); err != nil {
			logger.Log.Warningf(uiUtils.Warning, err)
		}
		return
	}

	logger.Log.Infof("test")

	if f.Name == SetCommandName {
		if err := mergeSetFlag(configElemValue, sliceValue.GetSlice()); err != nil {
			logger.Log.Warningf(uiUtils.Warning, err)
		}
		return
	}

	if err := mergeFlagValues(configElemValue, flagPath, strings.Join(flagPath, "."), sliceValue.GetSlice()); err != nil {
		logger.Log.Warningf(uiUtils.Warning, err)
	}
}

func mergeSetFlag(configElemValue reflect.Value, setValues []string) error {
	var setErrors []string
	setMap := map[string][]string{}

	for _, setValue := range setValues {
		if !strings.Contains(setValue, Separator) {
			setErrors = append(setErrors, fmt.Sprintf("Ignoring set argument %s (set argument format: <flag name>=<flag value>)", setValue))
			continue
		}

		split := strings.SplitN(setValue, Separator, 2)
		argumentKey, argumentValue := split[0], split[1]

		setMap[argumentKey] = append(setMap[argumentKey], argumentValue)
	}

	for argumentKey, argumentValues := range setMap {
		flagPath := strings.Split(argumentKey, ".")

		if len(argumentValues) > 1 {
			if err := mergeFlagValues(configElemValue, flagPath, argumentKey, argumentValues); err != nil {
				setErrors = append(setErrors, fmt.Sprintf("%v", err))
			}
		} else {
			if err := mergeFlagValue(configElemValue, flagPath, argumentKey, argumentValues[0]); err != nil {
				setErrors = append(setErrors, fmt.Sprintf("%v", err))
			}
		}
	}

	if len(setErrors) > 0 {
		return fmt.Errorf(strings.Join(setErrors, "\n"))
	}

	return nil
}

func mergeFlagValue(configElemValue reflect.Value, flagPath []string, fullFlagName string, flagValue string) error {
	mergeFunction := func(flagName string, currentFieldStruct reflect.StructField, currentFieldElemValue reflect.Value, currentElemValue reflect.Value) error {
		currentFieldKind := currentFieldStruct.Type.Kind()

		if currentFieldKind == reflect.Slice {
			return mergeFlagValues(currentElemValue, []string{flagName}, fullFlagName, []string{flagValue})
		}

		parsedValue, err := getParsedValue(currentFieldKind, flagValue)
		if err != nil {
			return fmt.Errorf("invalid value %s for flag name %s, expected %s", flagValue, flagName, currentFieldKind)
		}

		currentFieldElemValue.Set(parsedValue)
		return nil
	}

	return mergeFlag(configElemValue, flagPath, fullFlagName, mergeFunction)
}

func mergeFlagValues(configElemValue reflect.Value, flagPath []string, fullFlagName string, flagValues []string) error {
	mergeFunction := func(flagName string, currentFieldStruct reflect.StructField, currentFieldElemValue reflect.Value, currentElemValue reflect.Value) error {
		currentFieldKind := currentFieldStruct.Type.Kind()

		if currentFieldKind != reflect.Slice {
			return fmt.Errorf("invalid values %s for flag name %s, expected %s", strings.Join(flagValues, ","), flagName, currentFieldKind)
		}

		flagValueKind := currentFieldStruct.Type.Elem().Kind()

		parsedValues := reflect.MakeSlice(reflect.SliceOf(currentFieldStruct.Type.Elem()), 0, 0)
		for _, flagValue := range flagValues {
			parsedValue, err := getParsedValue(flagValueKind, flagValue)
			if err != nil {
				return fmt.Errorf("invalid value %s for flag name %s, expected %s", flagValue, flagName, flagValueKind)
			}

			parsedValues = reflect.Append(parsedValues, parsedValue)
		}

		currentFieldElemValue.Set(parsedValues)
		return nil
	}

	return mergeFlag(configElemValue, flagPath, fullFlagName, mergeFunction)
}

func mergeFlag(currentElemValue reflect.Value, currentFlagPath []string, fullFlagName string, mergeFunction func(flagName string, currentFieldStruct reflect.StructField, currentFieldElemValue reflect.Value, currentElemValue reflect.Value) error) error {
	if len(currentFlagPath) == 0 {
		return fmt.Errorf("flag \"%s\" not found", fullFlagName)
	}

	for i := 0; i < currentElemValue.NumField(); i++ {
		currentFieldStruct := currentElemValue.Type().Field(i)
		currentFieldElemValue := currentElemValue.FieldByName(currentFieldStruct.Name)

		if currentFieldStruct.Type.Kind() == reflect.Struct && getFieldNameByTag(currentFieldStruct) == currentFlagPath[0] {
			return mergeFlag(currentFieldElemValue, currentFlagPath[1:], fullFlagName, mergeFunction)
		}

		if len(currentFlagPath) > 1 || getFieldNameByTag(currentFieldStruct) != currentFlagPath[0] {
			continue
		}

		return mergeFunction(currentFlagPath[0], currentFieldStruct, currentFieldElemValue, currentElemValue)
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
