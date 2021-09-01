package config

import (
	"fmt"
	"github.com/up9inc/mizu/cli/logger"
	"github.com/up9inc/mizu/cli/mizu"
	"io/ioutil"
	"os"
	"reflect"
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
	Config  = ConfigStruct{}
	cmdName string
)

func InitConfig(cmd *cobra.Command) error {
	cmdName = cmd.Name()

	if err := defaults.Set(&Config); err != nil {
		return err
	}

	configFilePathFlag := cmd.Flags().Lookup(ConfigFilePathCommandName)
	configFilePath := configFilePathFlag.Value.String()
	if err := mergeConfigFile(configFilePath); err != nil {
		if configFilePathFlag.Changed || !os.IsNotExist(err) {
			return fmt.Errorf("invalid config, %w\n"+
				"you can regenerate the file by removing it (%v) and using `mizu config -r`", err, configFilePath)
		}
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

func mergeConfigFile(configFilePath string) error {
	reader, openErr := os.Open(configFilePath)
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
	logger.Log.Debugf("Found config file, merged to default options")

	return nil
}

func initFlag(f *pflag.Flag) {
	configElemValue := reflect.ValueOf(&Config).Elem()

	var flagPath []string
	if mizu.Contains([]string{ConfigFilePathCommandName}, f.Name) {
		flagPath = []string{f.Name}
	} else {
		flagPath = []string{cmdName, f.Name}
	}

	sliceValue, isSliceValue := f.Value.(pflag.SliceValue)
	if !isSliceValue {
		if err := mergeFlagValue(configElemValue, flagPath, strings.Join(flagPath, "."), f.Value.String()); err != nil {
			logger.Log.Warningf(uiUtils.Warning, err)
		}
		return
	}

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
