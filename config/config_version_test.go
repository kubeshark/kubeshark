package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/kubeshark/kubeshark/config/configStructs"
)

func TestValidateConfigVersion(t *testing.T) {
	tests := []struct {
		name          string
		version       int
		wantErrSubstr string
	}{
		{
			name:    "current version",
			version: currentConfigVersion,
		},
		{
			name:          "older version",
			version:       0,
			wantErrSubstr: "this CLI does not support config version",
		},
		{
			name:          "newer version",
			version:       currentConfigVersion + 1,
			wantErrSubstr: "this CLI does not support config version",
		},
		{
			name:          "negative version",
			version:       -1,
			wantErrSubstr: "this CLI does not support config version",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertError(t, validateConfigVersion(test.version), test.wantErrSubstr)
		})
	}
}

func TestConfigConstructorsUseCurrentVersion(t *testing.T) {
	if got := CreateDefaultConfig().Version; got != currentConfigVersion {
		t.Fatalf("CreateDefaultConfig().Version = %d, want %d", got, currentConfigVersion)
	}

	config, err := GetConfigWithDefaults()
	if err != nil {
		t.Fatalf("GetConfigWithDefaults() error: %v", err)
	}
	if config.Version != currentConfigVersion {
		t.Fatalf("GetConfigWithDefaults().Version = %d, want %d", config.Version, currentConfigVersion)
	}
}

func TestWriteConfigUsesCurrentVersion(t *testing.T) {
	setConfigFilePathForTest(t, filepath.Join(t.TempDir(), "config.yaml"))

	config := ConfigStruct{Version: currentConfigVersion + 1}
	if err := WriteConfig(&config); err != nil {
		t.Fatalf("WriteConfig() error: %v", err)
	}

	data, err := os.ReadFile(ConfigFilePath)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	if !strings.HasPrefix(string(data), fmt.Sprintf("version: %d\n", currentConfigVersion)) {
		t.Fatalf("written config does not start with current version:\n%s", data)
	}
	if config.Version != currentConfigVersion+1 {
		t.Fatalf("WriteConfig() mutated its input version: got %d", config.Version)
	}
}

func TestHelmValuesUseCurrentConfigVersion(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "helm-chart", "values.yaml"))
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	if _, err := decodeConfig(data, ConfigStruct{}); err != nil {
		t.Fatalf("helm values config version is out of sync: %v", err)
	}
}

func TestLoadConfigFile(t *testing.T) {
	tests := []struct {
		name          string
		data          string
		wantErrSubstr string
		wantLogLevel  string
	}{
		{
			name:         "legacy file",
			data:         "logLevel: debug\n",
			wantLogLevel: "debug",
		},
		{
			name:         "current file",
			data:         fmt.Sprintf("version: %d\nlogLevel: debug\n", currentConfigVersion),
			wantLogLevel: "debug",
		},
		{
			name:          "incompatible file",
			data:          fmt.Sprintf("version: %d\nlogLevel: debug\n", currentConfigVersion+1),
			wantErrSubstr: "this CLI does not support config version",
			wantLogLevel:  "warning",
		},
		{
			name:          "non-integer version",
			data:          "version: invalid\nlogLevel: debug\n",
			wantErrSubstr: "cannot unmarshal",
			wantLogLevel:  "warning",
		},
		{
			name:          "malformed file",
			data:          "version: [\nlogLevel: debug\n",
			wantErrSubstr: "cannot unmarshal",
			wantLogLevel:  "warning",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setConfigFilePathForTest(t, filepath.Join(t.TempDir(), "config.yaml"))
			if err := os.WriteFile(ConfigFilePath, []byte(test.data), 0o600); err != nil {
				t.Fatalf("WriteFile() error: %v", err)
			}

			loaded := ConfigStruct{Version: currentConfigVersion, LogLevel: "warning"}
			beforeLoad := loaded
			err := loadConfigFile(&loaded, true)
			assertError(t, err, test.wantErrSubstr)
			if test.wantErrSubstr == "" {
				if loaded.Version != currentConfigVersion {
					t.Fatalf("loaded version = %d, want %d", loaded.Version, currentConfigVersion)
				}
			} else {
				if !reflect.DeepEqual(loaded, beforeLoad) {
					t.Fatalf("loadConfigFile() changed config after an error: got %#v, want %#v", loaded, beforeLoad)
				}
			}

			if loaded.LogLevel != test.wantLogLevel {
				t.Fatalf("loaded log level = %q, want %q", loaded.LogLevel, test.wantLogLevel)
			}
		})
	}
}

func TestLoadConfigFileForCommand(t *testing.T) {
	tests := []struct {
		name          string
		command       func(*testing.T) *cobra.Command
		data          string
		wantErrSubstr string
		wantLogLevel  string
	}{
		{
			name: "ordinary command loads config",
			command: func(*testing.T) *cobra.Command {
				return &cobra.Command{Use: "tap"}
			},
			data:         "logLevel: debug\n",
			wantLogLevel: "debug",
		},
		{
			name: "config output loads config",
			command: func(t *testing.T) *cobra.Command {
				return newConfigCommandForTest(t, false)
			},
			data:         "logLevel: debug\n",
			wantLogLevel: "debug",
		},
		{
			name: "regeneration bypasses malformed config",
			command: func(t *testing.T) *cobra.Command {
				return newConfigCommandForTest(t, true)
			},
			data:         "version: [\n",
			wantLogLevel: "warning",
		},
		{
			name: "regeneration bypasses incompatible config",
			command: func(t *testing.T) *cobra.Command {
				return newConfigCommandForTest(t, true)
			},
			data:         fmt.Sprintf("version: %d\nlogLevel: debug\n", currentConfigVersion+1),
			wantLogLevel: "warning",
		},
		{
			name: "config command requires regeneration flag",
			command: func(*testing.T) *cobra.Command {
				return &cobra.Command{Use: "config"}
			},
			data:          "logLevel: debug\n",
			wantErrSubstr: "failed reading --regenerate flag",
			wantLogLevel:  "warning",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setConfigFilePathForTest(t, filepath.Join(t.TempDir(), "config.yaml"))
			if err := os.WriteFile(ConfigFilePath, []byte(test.data), 0o600); err != nil {
				t.Fatalf("WriteFile() error: %v", err)
			}

			loaded := ConfigStruct{LogLevel: "warning"}
			assertError(t, loadConfigFileForCommand(&loaded, test.command(t), true), test.wantErrSubstr)

			if loaded.LogLevel != test.wantLogLevel {
				t.Fatalf("loaded log level = %q, want %q", loaded.LogLevel, test.wantLogLevel)
			}
		})
	}
}

func TestInitConfigReportsIncompatibleVersion(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	data := fmt.Sprintf("version: %d\n", currentConfigVersion+1)
	if err := os.WriteFile(configPath, []byte(data), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	originalConfig := Config
	originalDebugMode := DebugMode
	originalCommandName := cmdName
	originalPath := ConfigFilePath
	t.Cleanup(func() {
		Config = originalConfig
		DebugMode = originalDebugMode
		cmdName = originalCommandName
		ConfigFilePath = originalPath
	})

	cmd := &cobra.Command{Use: "license"}
	cmd.Flags().Bool(DebugFlag, false, "")
	cmd.Flags().String(ConfigPathFlag, configPath, "")

	err := InitConfig(cmd)
	if err == nil {
		t.Fatal("InitConfig() expected an incompatibility error")
	}
	for _, expected := range []string{
		"this CLI does not support config version",
		"kubeshark config -r",
	} {
		if !strings.Contains(err.Error(), expected) {
			t.Fatalf("InitConfig() error = %q, want %q", err, expected)
		}
	}
}

func newConfigCommandForTest(t *testing.T, regenerate bool) *cobra.Command {
	t.Helper()

	cmd := &cobra.Command{Use: "config"}
	cmd.Flags().Bool(configStructs.RegenerateConfigName, false, "")
	if regenerate {
		if err := cmd.Flags().Set(configStructs.RegenerateConfigName, "true"); err != nil {
			t.Fatalf("Set(%q) error: %v", configStructs.RegenerateConfigName, err)
		}
	}

	return cmd
}

func setConfigFilePathForTest(t *testing.T, path string) {
	t.Helper()

	originalPath := ConfigFilePath
	ConfigFilePath = path
	t.Cleanup(func() {
		ConfigFilePath = originalPath
	})
}

func assertError(t *testing.T, err error, wantSubstring string) {
	t.Helper()

	if wantSubstring == "" {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return
	}

	if err == nil {
		t.Fatalf("expected an error containing %q", wantSubstring)
	}
	if !strings.Contains(err.Error(), wantSubstring) {
		t.Fatalf("error = %q, want substring %q", err, wantSubstring)
	}
}
