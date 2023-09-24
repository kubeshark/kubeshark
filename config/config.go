package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/creasty/defaults"
	"github.com/goccy/go-yaml"
	"github.com/kubeshark/kubeshark/misc"
	"github.com/kubeshark/kubeshark/misc/version"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	Separator      = "="
	SetCommandName = "set"
	FieldNameTag   = "yaml"
	ReadonlyTag    = "readonly"
	DebugFlag      = "debug"
)

var (
	Config         ConfigStruct
	DebugMode      bool
	cmdName        string
	ConfigFilePath string
)

func InitConfig(cmd *cobra.Command) error {
	var err error
	DebugMode, err = cmd.Flags().GetBool(DebugFlag)
	if err != nil {
		log.Error().Err(err).Msg(fmt.Sprintf("Can't recieve '%s' flag", DebugFlag))
	}

	if DebugMode {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if cmd.Use == "version" {
		return nil
	}

	if !utils.Contains([]string{
		"console",
		"pro",
		"manifests",
		"license",
	}, cmd.Use) {
		go version.CheckNewerVersion()
	}

	Config = CreateDefaultConfig()
	Config.Tap.Debug = DebugMode
	cmdName = cmd.Name()
	if utils.Contains([]string{
		"clean",
		"console",
		"pro",
		"proxy",
		"scripts",
	}, cmdName) {
		cmdName = "tap"
	}

	if err := defaults.Set(&Config); err != nil {
		return err
	}

	ConfigFilePath = path.Join(misc.GetDotFolderPath(), "config.yaml")
	if err := loadConfigFile(&Config, utils.Contains([]string{
		"manifests",
		"license",
	}, cmd.Use)); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("invalid config, %w\n"+
				"you can regenerate the file by removing it (%v) and using `kubeshark config -r`", err, ConfigFilePath)
		}
	}

	cmd.Flags().Visit(initFlag)

	log.Debug().Interface("config", Config).Msg("Init config is finished.")

	return nil
}

func GetConfigWithDefaults() (*ConfigStruct, error) {
	defaultConf := ConfigStruct{}
	if err := defaults.Set(&defaultConf); err != nil {
		return nil, err
	}

	configElem := reflect.ValueOf(&defaultConf).Elem()
	setZeroForReadonlyFields(configElem)

	return &defaultConf, nil
}

func WriteConfig(config *ConfigStruct) error {
	template, err := utils.PrettyYaml(config)
	if err != nil {
		return fmt.Errorf("failed converting config to yaml, err: %v", err)
	}

	data := []byte(template)

	if _, err := os.Stat(ConfigFilePath); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(ConfigFilePath), 0700)
		if err != nil {
			return fmt.Errorf("failed creating directories, err: %v", err)
		}
	}

	if err := os.WriteFile(ConfigFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed writing config, err: %v", err)
	}

	return nil
}

func loadConfigFile(config *ConfigStruct, silent bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	cwdConfig := filepath.Join(cwd, fmt.Sprintf("%s.yaml", misc.Program))
	reader, err := os.Open(cwdConfig)
	if err != nil {
		reader, err = os.Open(ConfigFilePath)
		if err != nil {
			return err
		}
	} else {
		ConfigFilePath = cwdConfig
	}

	buf, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(buf, config); err != nil {
		return err
	}

	if !silent {
		log.Info().Str("path", ConfigFilePath).Msg("Found config file!")
	}

	return nil
}

func initFlag(f *pflag.Flag) {
	configElemValue := reflect.ValueOf(&Config).Elem()

	var flagPath []string
	flagPath = append(flagPath, cmdName)

	flagPath = append(flagPath, strings.Split(f.Name, "-")...)

	sliceValue, isSliceValue := f.Value.(pflag.SliceValue)
	if !isSliceValue {
		if err := mergeFlagValue(configElemValue, flagPath, strings.Join(flagPath, "."), f.Value.String()); err != nil {
			log.Warn().Err(err).Send()
		}
		return
	}

	if f.Name == SetCommandName {
		if err := mergeSetFlag(configElemValue, sliceValue.GetSlice()); err != nil {
			log.Warn().Err(err).Send()
		}
		return
	}

	if err := mergeFlagValues(configElemValue, flagPath, strings.Join(flagPath, "."), sliceValue.GetSlice()); err != nil {
		log.Warn().Err(err).Send()
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
