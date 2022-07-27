package acceptanceTests

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"gopkg.in/yaml.v3"
)

type tapConfig struct {
	GuiPort uint16 `yaml:"gui-port"`
}

type configStruct struct {
	Tap tapConfig `yaml:"tap"`
}

func TestConfigRegenerate(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	cliPath, cliPathErr := GetCliPath()
	if cliPathErr != nil {
		t.Errorf("failed to get cli path, err: %v", cliPathErr)
		return
	}

	configPath, configPathErr := GetConfigPath()
	if configPathErr != nil {
		t.Errorf("failed to get config path, err: %v", cliPathErr)
		return
	}

	configCmdArgs := GetDefaultConfigCommandArgs()

	configCmdArgs = append(configCmdArgs, "-r")

	configCmd := exec.Command(cliPath, configCmdArgs...)
	t.Logf("running command: %v", configCmd.String())

	t.Cleanup(func() {
		if err := os.Remove(configPath); err != nil {
			t.Logf("failed to delete config file, err: %v", err)
		}
	})

	if err := configCmd.Start(); err != nil {
		t.Errorf("failed to start config command, err: %v", err)
		return
	}

	if err := configCmd.Wait(); err != nil {
		t.Errorf("failed to wait config command, err: %v", err)
		return
	}

	_, readFileErr := ioutil.ReadFile(configPath)
	if readFileErr != nil {
		t.Errorf("failed to read config file, err: %v", readFileErr)
		return
	}
}

func TestConfigGuiPort(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	tests := []uint16{8898}

	for _, guiPort := range tests {
		t.Run(fmt.Sprintf("%d", guiPort), func(t *testing.T) {
			cliPath, cliPathErr := GetCliPath()
			if cliPathErr != nil {
				t.Errorf("failed to get cli path, err: %v", cliPathErr)
				return
			}

			configPath, configPathErr := GetConfigPath()
			if configPathErr != nil {
				t.Errorf("failed to get config path, err: %v", cliPathErr)
				return
			}

			config := configStruct{}
			config.Tap.GuiPort = guiPort

			configBytes, marshalErr := yaml.Marshal(config)
			if marshalErr != nil {
				t.Errorf("failed to marshal config, err: %v", marshalErr)
				return
			}

			if writeErr := ioutil.WriteFile(configPath, configBytes, 0644); writeErr != nil {
				t.Errorf("failed to write config to file, err: %v", writeErr)
				return
			}

			tapCmdArgs := GetDefaultTapCommandArgs()

			tapNamespace := GetDefaultTapNamespace()
			tapCmdArgs = append(tapCmdArgs, tapNamespace...)

			tapCmd := exec.Command(cliPath, tapCmdArgs...)
			t.Logf("running command: %v", tapCmd.String())

			t.Cleanup(func() {
				if err := CleanupCommand(tapCmd); err != nil {
					t.Logf("failed to cleanup tap command, err: %v", err)
				}

				if err := os.Remove(configPath); err != nil {
					t.Logf("failed to delete config file, err: %v", err)
				}
			})

			if err := tapCmd.Start(); err != nil {
				t.Errorf("failed to start tap command, err: %v", err)
				return
			}

			apiServerUrl := GetApiServerUrl(guiPort)

			if err := WaitTapPodsReady(apiServerUrl); err != nil {
				t.Errorf("failed to start tap pods on time, err: %v", err)
				return
			}
		})
	}
}

func TestConfigSetGuiPort(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	tests := []struct {
		ConfigFileGuiPort uint16
		SetGuiPort        uint16
	}{
		{ConfigFileGuiPort: 8898, SetGuiPort: 8897},
	}

	for _, guiPortStruct := range tests {
		t.Run(fmt.Sprintf("%d", guiPortStruct.SetGuiPort), func(t *testing.T) {
			cliPath, cliPathErr := GetCliPath()
			if cliPathErr != nil {
				t.Errorf("failed to get cli path, err: %v", cliPathErr)
				return
			}

			configPath, configPathErr := GetConfigPath()
			if configPathErr != nil {
				t.Errorf("failed to get config path, err: %v", cliPathErr)
				return
			}

			config := configStruct{}
			config.Tap.GuiPort = guiPortStruct.ConfigFileGuiPort

			configBytes, marshalErr := yaml.Marshal(config)
			if marshalErr != nil {
				t.Errorf("failed to marshal config, err: %v", marshalErr)
				return
			}

			if writeErr := ioutil.WriteFile(configPath, configBytes, 0644); writeErr != nil {
				t.Errorf("failed to write config to file, err: %v", writeErr)
				return
			}

			tapCmdArgs := GetDefaultTapCommandArgs()

			tapNamespace := GetDefaultTapNamespace()
			tapCmdArgs = append(tapCmdArgs, tapNamespace...)

			tapCmdArgs = append(tapCmdArgs, "--set", fmt.Sprintf("tap.gui-port=%v", guiPortStruct.SetGuiPort))

			tapCmd := exec.Command(cliPath, tapCmdArgs...)
			t.Logf("running command: %v", tapCmd.String())

			t.Cleanup(func() {
				if err := CleanupCommand(tapCmd); err != nil {
					t.Logf("failed to cleanup tap command, err: %v", err)
				}

				if err := os.Remove(configPath); err != nil {
					t.Logf("failed to delete config file, err: %v", err)
				}
			})

			if err := tapCmd.Start(); err != nil {
				t.Errorf("failed to start tap command, err: %v", err)
				return
			}

			apiServerUrl := GetApiServerUrl(guiPortStruct.SetGuiPort)

			if err := WaitTapPodsReady(apiServerUrl); err != nil {
				t.Errorf("failed to start tap pods on time, err: %v", err)
				return
			}
		})
	}
}

func TestConfigFlagGuiPort(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	tests := []struct {
		ConfigFileGuiPort uint16
		FlagGuiPort       uint16
	}{
		{ConfigFileGuiPort: 8898, FlagGuiPort: 8896},
	}

	for _, guiPortStruct := range tests {
		t.Run(fmt.Sprintf("%d", guiPortStruct.FlagGuiPort), func(t *testing.T) {
			cliPath, cliPathErr := GetCliPath()
			if cliPathErr != nil {
				t.Errorf("failed to get cli path, err: %v", cliPathErr)
				return
			}

			configPath, configPathErr := GetConfigPath()
			if configPathErr != nil {
				t.Errorf("failed to get config path, err: %v", cliPathErr)
				return
			}

			config := configStruct{}
			config.Tap.GuiPort = guiPortStruct.ConfigFileGuiPort

			configBytes, marshalErr := yaml.Marshal(config)
			if marshalErr != nil {
				t.Errorf("failed to marshal config, err: %v", marshalErr)
				return
			}

			if writeErr := ioutil.WriteFile(configPath, configBytes, 0644); writeErr != nil {
				t.Errorf("failed to write config to file, err: %v", writeErr)
				return
			}

			tapCmdArgs := GetDefaultTapCommandArgs()

			tapNamespace := GetDefaultTapNamespace()
			tapCmdArgs = append(tapCmdArgs, tapNamespace...)

			tapCmdArgs = append(tapCmdArgs, "-p", fmt.Sprintf("%v", guiPortStruct.FlagGuiPort))

			tapCmd := exec.Command(cliPath, tapCmdArgs...)
			t.Logf("running command: %v", tapCmd.String())

			t.Cleanup(func() {
				if err := CleanupCommand(tapCmd); err != nil {
					t.Logf("failed to cleanup tap command, err: %v", err)
				}

				if err := os.Remove(configPath); err != nil {
					t.Logf("failed to delete config file, err: %v", err)
				}
			})

			if err := tapCmd.Start(); err != nil {
				t.Errorf("failed to start tap command, err: %v", err)
				return
			}

			apiServerUrl := GetApiServerUrl(guiPortStruct.FlagGuiPort)

			if err := WaitTapPodsReady(apiServerUrl); err != nil {
				t.Errorf("failed to start tap pods on time, err: %v", err)
				return
			}
		})
	}
}
