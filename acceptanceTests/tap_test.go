package acceptanceTests

import (
	"archive/zip"
	"io/ioutil"
	"os/exec"
	"path"
	"strings"
	"testing"
)

func TestTapDumpLogs(t *testing.T) {
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

	tapCmdArgs = append(tapCmdArgs, "--set", "dump-logs=true")

	tapCmd := exec.Command(cliPath, tapCmdArgs...)
	t.Logf("running command: %v", tapCmd.String())

	if err := tapCmd.Start(); err != nil {
		t.Errorf("failed to start tap command, err: %v", err)
		return
	}

	apiServerUrl := getApiServerUrl(defaultApiServerPort)

	if err := waitTapPodsReady(apiServerUrl); err != nil {
		t.Errorf("failed to start tap pods on time, err: %v", err)
		return
	}

	if err := cleanupCommand(tapCmd); err != nil {
		t.Errorf("failed to cleanup tap command, err: %v", err)
		return
	}

	mizuFolderPath, mizuPathErr := getMizuFolderPath()
	if mizuPathErr != nil {
		t.Errorf("failed to get mizu folder path, err: %v", mizuPathErr)
		return
	}

	files, readErr := ioutil.ReadDir(mizuFolderPath)
	if readErr != nil {
		t.Errorf("failed to read mizu folder files, err: %v", readErr)
		return
	}

	var dumpLogsPath string
	for _, file := range files {
		fileName := file.Name()
		if strings.Contains(fileName, "mizu_logs") {
			dumpLogsPath = path.Join(mizuFolderPath, fileName)
			break
		}
	}

	if dumpLogsPath == "" {
		t.Errorf("dump logs file not found")
		return
	}

	zipReader, zipError := zip.OpenReader(dumpLogsPath)
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

	for _, file := range zipReader.File {
		if file.Name == "mizu_cli.log" {
			fc, _ := file.Open()
			content, _ := ioutil.ReadAll(fc)
			t.Logf("%s", content)
		}
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
