package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/kubernetes"
	"github.com/up9inc/mizu/cli/logger"
	"github.com/up9inc/mizu/cli/mizu"
	"io/ioutil"
	"net/http"
)

const telemetryUrl = "https://us-east4-up9-prod.cloudfunctions.net/mizu-telemetry"

func ReportRun(cmd string, args interface{}) {
	if !shouldRunTelemetry() {
		logger.Log.Debugf("not reporting telemetry")
	}

	argsBytes, _ := json.Marshal(args)
	argsMap := map[string]interface{}{
		"telemetry_type": "execution",
		"cmd":            cmd,
		"args":           string(argsBytes),
		"component":      "mizu_cli",
		"BuildTimestamp": mizu.BuildTimestamp,
		"Branch":         mizu.Branch,
		"version":        mizu.SemVer,
	}

	if err := sentTelemetry(argsMap); err != nil {
		logger.Log.Debugf("%v", err)
		return
	}

	logger.Log.Debugf("successfully reported telemetry for cmd %v", cmd)
}

func ReportEntriesCount(mizuPort uint16) {
	if !shouldRunTelemetry() {
		logger.Log.Debugf("not reporting telemetry")
	}

	mizuProxiedUrl := kubernetes.GetMizuApiServerProxiedHostAndPath(mizuPort)
	generalStatsUrl := fmt.Sprintf("http://%s/api/generalStats", mizuProxiedUrl)

	response, requestErr := http.Get(generalStatsUrl)
	if requestErr != nil {
		logger.Log.Debugf("failed to get general stats for telemetry, err: %v", requestErr)
		return
	} else if response.StatusCode != 200 {
		logger.Log.Debugf("failed to get general stats for telemetry, status code: %v", response.StatusCode)
		return
	}

	defer func() { _ = response.Body.Close() }()

	data, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		logger.Log.Debugf("failed to read general stats for telemetry, err: %v", readErr)
		return
	}

	var generalStats map[string]interface{}
	if parseErr := json.Unmarshal(data, &generalStats); parseErr != nil {
		logger.Log.Debugf("failed to parse general stats for telemetry, err: %v", parseErr)
		return
	}

	argsMap := map[string]interface{}{
		"telemetry_type":        "APICalls",
		"component":             "mizu_cli",
		"Count":                 generalStats["EntriesCount"],
		"FirstAPICallTimestamp": generalStats["FirstEntryTimestamp"],
		"LastAPICallTimestamp":  generalStats["LastEntryTimestamp"],
		"BuildTimestamp":        mizu.BuildTimestamp,
		"Branch":                mizu.Branch,
		"version":               mizu.SemVer,
	}

	if err := sentTelemetry(argsMap); err != nil {
		logger.Log.Debugf("%v", err)
		return
	}

	logger.Log.Debugf("successfully reported telemetry of api calls")
}

func shouldRunTelemetry() bool {
	if !config.Config.Telemetry {
		return false
	}

	if mizu.Branch != "main" && mizu.Branch != "develop" {
		return false
	}

	return true
}

func sentTelemetry(argsMap map[string]interface{}) error {
	jsonValue, _ := json.Marshal(argsMap)

	if resp, err := http.Post(telemetryUrl, "application/json", bytes.NewBuffer(jsonValue)); err != nil {
		return fmt.Errorf("error sending telemetry err: %v, response %v", err, resp)
	}

	return nil
}
