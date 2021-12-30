package acceptanceTests

import (
	"fmt"
	"os/exec"
	"testing"
)

func TestTap(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	tests := []int{50}

	for _, entriesCount := range tests {
		t.Run(fmt.Sprintf("%d", entriesCount), func(t *testing.T) {
			cliPath, cliPathErr := getCliPath()
			if cliPathErr != nil {
				t.Errorf("failed to get cli path, err: %v", cliPathErr)
				return
			}

			tapCmdArgs := getDefaultTapCommandArgs()

			tapNamespace := getDefaultTapNamespace()
			tapCmdArgs = append(tapCmdArgs, tapNamespace...)

			tapCmd := exec.Command(cliPath, tapCmdArgs...)
			t.Logf("running command: %v", tapCmd.String())

			t.Cleanup(func() {
				if err := cleanupCommand(tapCmd); err != nil {
					t.Logf("failed to cleanup tap command, err: %v", err)
				}
			})

			if err := tapCmd.Start(); err != nil {
				t.Errorf("failed to start tap command, err: %v", err)
				return
			}

			apiServerUrl := getApiServerUrl(defaultApiServerPort)

			if err := waitTapPodsReady(apiServerUrl); err != nil {
				t.Errorf("failed to start tap pods on time, err: %v", err)
				return
			}

			proxyUrl := getProxyUrl(defaultNamespaceName, defaultServiceName)
			for i := 0; i < entriesCount; i++ {
				if _, requestErr := executeHttpGetRequest(fmt.Sprintf("%v/get", proxyUrl)); requestErr != nil {
					t.Errorf("failed to send proxy request, err: %v", requestErr)
					return
				}
			}

			cypressCmd := exec.Command("bash", "-c", "npx cypress run")
			t.Logf("running command: %v", cypressCmd.String())
			out, err := cypressCmd.Output()
			if err != nil {
				t.Errorf("%s", out)
				return
			}
			t.Logf("%s", out)
		})
	}
}
