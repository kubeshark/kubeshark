package acceptanceTests

import (
	"archive/zip"
	"os/exec"
	"testing"
)

func TestLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

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

	logsCmdArgs := getDefaultLogsCommandArgs()

	logsCmd := exec.Command(cliPath, logsCmdArgs...)
	t.Logf("running command: %v", logsCmd.String())

	if err := logsCmd.Start(); err != nil {
		t.Errorf("failed to start logs command, err: %v", err)
		return
	}

	if err := logsCmd.Wait(); err != nil {
		t.Errorf("failed to wait logs command, err: %v", err)
		return
	}

	logsPath, logsPathErr := getLogsPath()
	if logsPathErr != nil {
		t.Errorf("failed to get logs path, err: %v", logsPathErr)
		return
	}

	zipReader, zipError := zip.OpenReader(logsPath)
	if zipError != nil {
		t.Errorf("failed to get zip reader, err: %v", zipError)
		return
	}

	t.Cleanup(func() {
		if err := zipReader.Close(); err != nil {
			t.Logf("failed to close zip reader, err: %v", err)
		}
	})

	var logsFileNames []string
	for _, file := range zipReader.File {
		logsFileNames = append(logsFileNames, file.Name)
	}

	if !Contains(logsFileNames, "mizu.mizu-api-server.mizu-api-server.log") {
		t.Errorf("api server logs not found")
		return
	}

	if !Contains(logsFileNames, "mizu.mizu-api-server.basenine.log") {
		t.Errorf("basenine logs not found")
		return
	}

	if !Contains(logsFileNames, "mizu_cli.log") {
		t.Errorf("cli logs not found")
		return
	}

	if !Contains(logsFileNames, "mizu_events.log") {
		t.Errorf("events logs not found")
		return
	}

	if !ContainsPartOfValue(logsFileNames, "mizu.mizu-tapper-daemon-set") {
		t.Errorf("tapper logs not found")
		return
	}
}

func TestLogsPath(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

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

	logsCmdArgs := getDefaultLogsCommandArgs()

	logsPath := "../logs.zip"
	logsCmdArgs = append(logsCmdArgs, "-f", logsPath)

	logsCmd := exec.Command(cliPath, logsCmdArgs...)
	t.Logf("running command: %v", logsCmd.String())

	if err := logsCmd.Start(); err != nil {
		t.Errorf("failed to start logs command, err: %v", err)
		return
	}

	if err := logsCmd.Wait(); err != nil {
		t.Errorf("failed to wait logs command, err: %v", err)
		return
	}

	zipReader, zipError := zip.OpenReader(logsPath)
	if zipError != nil {
		t.Errorf("failed to get zip reader, err: %v", zipError)
		return
	}

	t.Cleanup(func() {
		if err := zipReader.Close(); err != nil {
			t.Logf("failed to close zip reader, err: %v", err)
		}
	})

	var logsFileNames []string
	for _, file := range zipReader.File {
		logsFileNames = append(logsFileNames, file.Name)
	}

	if !Contains(logsFileNames, "mizu.mizu-api-server.log") {
		t.Errorf("api server logs not found")
		return
	}

	if !Contains(logsFileNames, "mizu_cli.log") {
		t.Errorf("cli logs not found")
		return
	}

	if !Contains(logsFileNames, "mizu_events.log") {
		t.Errorf("events logs not found")
		return
	}

	if !ContainsPartOfValue(logsFileNames, "mizu.mizu-tapper-daemon-set") {
		t.Errorf("tapper logs not found")
		return
	}
}
