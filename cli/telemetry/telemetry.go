package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/up9inc/mizu/cli/apiserver"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/shared/logger"
)

const telemetryUrl = "https://us-east4-up9-prod.cloudfunctions.net/mizu-telemetry"

var runStartTime time.Time

func ReportRun(cmd string, args interface{}) {
	if !shouldRunTelemetry() {
		logger.Log.Debugf("not reporting telemetry")
		return
	}

	runStartTime = time.Now()

	argsBytes, _ := json.Marshal(args)
	argsMap := map[string]interface{}{
		"cmd":  cmd,
		"args": string(argsBytes),
	}

	if err := sendTelemetry("Execution", argsMap); err != nil {
		logger.Log.Debug(err)
		return
	}

	logger.Log.Debugf("successfully reported telemetry for cmd %v", cmd)
}

func ReportExecutionTime() {
	if !shouldRunTelemetry() {
		logger.Log.Debugf("not reporting telemetry")
		return
	}

	executionTime := getExecutionTime()
	argsMap := map[string]interface{}{
		"timeInSeconds": executionTime.Seconds(),
	}

	if err := sendTelemetry("ExecutionTime", argsMap); err != nil {
		logger.Log.Debug(err)
		return
	}

	logger.Log.Debugf("successfully reported telemetry for execution time %v", executionTime.String())
}

func ReportAPICalls(apiProvider *apiserver.Provider) {
	if !shouldRunTelemetry() {
		logger.Log.Debugf("not reporting telemetry")
		return
	}

	generalStats, err := apiProvider.GetGeneralStats()
	if err != nil {
		logger.Log.Debugf("[ERROR] failed get general stats from api server %v", err)
		return
	}

	argsMap := map[string]interface{}{
		"apiCallsCount":         generalStats["EntriesCount"],
		"firstAPICallTimestamp": generalStats["FirstEntryTimestamp"],
		"lastAPICallTimestamp":  generalStats["LastEntryTimestamp"],
	}

	if err := sendTelemetry("APICalls", argsMap); err != nil {
		logger.Log.Debug(err)
		return
	}

	logger.Log.Debugf("successfully reported telemetry of api calls")
}

func getExecutionTime() time.Duration {
	return time.Since(runStartTime)
}

func shouldRunTelemetry() bool {
	if _, present := os.LookupEnv(mizu.DEVENVVAR); present {
		return false
	}
	if !config.Config.Telemetry {
		return false
	}

	if mizu.Branch != "main" && mizu.Branch != "develop" {
		return false
	}

	return true
}

func sendTelemetry(telemetryType string, argsMap map[string]interface{}) error {
	argsMap["telemetryType"] = telemetryType
	argsMap["component"] = "mizu_cli"
	argsMap["buildTimestamp"] = mizu.BuildTimestamp
	argsMap["branch"] = mizu.Branch
	argsMap["version"] = mizu.SemVer
	argsMap["Platform"] = mizu.Platform
	argsMap["executionTime"] = getExecutionTime().String()

	if machineId, err := machineid.ProtectedID("mizu"); err == nil {
		argsMap["machineId"] = machineId
	}

	jsonValue, _ := json.Marshal(argsMap)

	if resp, err := http.Post(telemetryUrl, "application/json", bytes.NewBuffer(jsonValue)); err != nil {
		return fmt.Errorf("ERROR: failed sending telemetry, err: %v, response %v", err, resp)
	}

	return nil
}
