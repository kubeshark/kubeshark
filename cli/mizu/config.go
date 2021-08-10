package mizu

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"

	"github.com/creasty/defaults"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/up9inc/mizu/cli/mizu/configStructs"
	"github.com/up9inc/mizu/cli/uiUtils"
	"gopkg.in/yaml.v3"
)

const (
	Separator      = "="
	SetCommandName = "set"
	FieldNameTag   = "yaml"
	ReadonlyTag    = "readonly"
)

var allowedSetFlags = []string{
	AgentImageConfigName,
	MizuResourcesNamespaceConfigName,
	TelemetryConfigName,
	DumpLogsConfigName,
	KubeConfigPathName,
	configStructs.AnalysisDestinationTapName,
	configStructs.SleepIntervalSecTapName,
	configStructs.IgnoredUserAgentsTapName,
}

var Config = ConfigStruct{}

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
	if err := defaults.Set(&Config); err != nil {
		return err
	}

	if err := mergeConfigFile(); err != nil {
		return fmt.Errorf("invalid config %w\n"+
			"you can regenerate the file using `mizu config -r` or just remove it %v", err, GetConfigFilePath())
	}

	cmd.Flags().Visit(initFlag)

	finalConfigPrettified, _ := uiUtils.PrettyJson(Config)
	Log.Debugf("Init config finished\n Final config: %v", finalConfigPrettified)

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
	return path.Join(GetMizuFolderPath(), "config.yaml")
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
	Log.Debugf("Found config file, merged to default options")

	return nil
}

func initFlag(f *pflag.Flag) {
	configElem := reflect.ValueOf(&Config).Elem()

	sliceValue, isSliceValue := f.Value.(pflag.SliceValue)
	if !isSliceValue {
		mergeFlagValue(configElem, f.Name, f.Value.String())
		return
	}

	if f.Name == SetCommandName {
		mergeSetFlag(configElem, sliceValue.GetSlice())
		return
	}

	mergeFlagValues(configElem, f.Name, sliceValue.GetSlice())
}

func mergeSetFlag(configElem reflect.Value, setValues []string) {
	setMap := map[string][]string{}

	for _, setValue := range setValues {
		if !strings.Contains(setValue, Separator) {
			Log.Warningf(uiUtils.Warning, fmt.Sprintf("Ignoring set argument %s (set argument format: <flag name>=<flag value>)", setValue))
			continue
		}

		split := strings.SplitN(setValue, Separator, 2)
		if len(split) != 2 {
			Log.Warningf(uiUtils.Warning, fmt.Sprintf("Ignoring set argument %s (set argument format: <flag name>=<flag value>)", setValue))
			continue
		}

		argumentKey, argumentValue := split[0], split[1]

		setMap[argumentKey] = append(setMap[argumentKey], argumentValue)
	}

	for argumentKey, argumentValues := range setMap {
		if !Contains(allowedSetFlags, argumentKey) {
			Log.Warningf(uiUtils.Warning, fmt.Sprintf("Ignoring set argument name \"%s\", flag name must be one of the following: \"%s\"", argumentKey, strings.Join(allowedSetFlags, "\", \"")))
			continue
		}

		if len(argumentValues) > 1 {
			mergeFlagValues(configElem, argumentKey, argumentValues)
		} else {
			mergeFlagValue(configElem, argumentKey, argumentValues[0])
		}
	}
}

func mergeFlagValue(currentElem reflect.Value, flagKey string, flagValue string) {
	for i := 0; i < currentElem.NumField(); i++ {
		currentField := currentElem.Type().Field(i)
		currentFieldByName := currentElem.FieldByName(currentField.Name)
		currentFieldKind := currentField.Type.Kind()

		if currentFieldKind == reflect.Struct {
			mergeFlagValue(currentFieldByName, flagKey, flagValue)
			continue
		}

		if getFieldNameByTag(currentField) != flagKey {
			continue
		}

		if currentFieldKind == reflect.Slice {
			mergeFlagValues(currentElem, flagKey, []string{flagValue})
			return
		}

		parsedValue, err := getParsedValue(currentFieldKind, flagValue)
		if err != nil {
			Log.Warningf(uiUtils.Warning, fmt.Sprintf("Invalid value %s for flag name %s, expected %s", flagValue, flagKey, currentFieldKind))
			return
		}

		currentFieldByName.Set(parsedValue)
	}
}

func mergeFlagValues(currentElem reflect.Value, flagKey string, flagValues []string) {
	for i := 0; i < currentElem.NumField(); i++ {
		currentField := currentElem.Type().Field(i)
		currentFieldByName := currentElem.FieldByName(currentField.Name)
		currentFieldKind := currentField.Type.Kind()

		if currentFieldKind == reflect.Struct {
			mergeFlagValues(currentFieldByName, flagKey, flagValues)
			continue
		}

		if getFieldNameByTag(currentField) != flagKey {
			continue
		}

		if currentFieldKind != reflect.Slice {
			Log.Warningf(uiUtils.Warning, fmt.Sprintf("Invalid values %s for flag name %s, expected %s", strings.Join(flagValues, ","), flagKey, currentFieldKind))
			return
		}

		flagValueKind := currentField.Type.Elem().Kind()

		parsedValues := reflect.MakeSlice(reflect.SliceOf(currentField.Type.Elem()), 0, 0)
		for _, flagValue := range flagValues {
			parsedValue, err := getParsedValue(flagValueKind, flagValue)
			if err != nil {
				Log.Warningf(uiUtils.Warning, fmt.Sprintf("Invalid value %s for flag name %s, expected %s", flagValue, flagKey, flagValueKind))
				return
			}

			parsedValues = reflect.Append(parsedValues, parsedValue)
		}

		currentFieldByName.Set(parsedValues)
	}
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
