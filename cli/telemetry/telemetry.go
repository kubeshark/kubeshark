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
	"github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/shared/logger"
)

const telemetryUrl = "https://us-east4-up9-prod.cloudfunctions.net/mizu-telemetry"

type telemetryType int

const (
	Execution telemetryType = iota
	TapExecution
)

func (t telemetryType) String() string {
	switch t {
	case Execution:
		return "Execution"
	case TapExecution:
		return "TapExecution"
	default:
		return "Unkown"
	}
}

type tapTelemetry struct {
	cmd       string
	args      string
	startTime time.Time
}

var tapTelemetryData *tapTelemetry

func ReportRun(cmd string, args interface{}) {
	if !shouldRunTelemetry() {
		logger.Log.Debug("not reporting telemetry")
		return
	}

	argsBytes, _ := json.Marshal(args)
	argsMap := map[string]interface{}{
		"cmd":  cmd,
		"args": string(argsBytes),
	}

	if err := sendTelemetry(Execution, argsMap); err != nil {
		logger.Log.Debug(err)
		return
	}

	logger.Log.Debugf("successfully reported telemetry for cmd %v", cmd)
}

func StartTapTelemetry(args configStructs.TapConfig) {
	argsBytes, _ := json.Marshal(args)
	tapTelemetryData = &tapTelemetry{
		cmd:       "tap",
		args:      string(argsBytes),
		startTime: time.Now(),
	}
}

func ReportTapTelemetry(apiProvider *apiserver.Provider) {
	if !shouldRunTelemetry() {
		logger.Log.Debug("not reporting telemetry")
		return
	}

	if tapTelemetryData == nil {
		logger.Log.Debug(`[ERROR] tap telemetry data is nil, you must call "StartTapTelemetry"`)
		return
	}

	generalStats, err := apiProvider.GetGeneralStats()
	if err != nil {
		logger.Log.Debugf("[ERROR] failed to get general stats from api server %v", err)
		return
	}

	argsMap := map[string]interface{}{
		"cmd":                    tapTelemetryData.cmd,
		"args":                   tapTelemetryData.args,
		"executionTimeInSeconds": getExecutionTime(tapTelemetryData.startTime).Seconds(),
		"apiCallsCount":          generalStats["EntriesCount"],
		"firstAPICallTimestamp":  generalStats["FirstEntryTimestamp"],
		"lastAPICallTimestamp":   generalStats["LastEntryTimestamp"],
	}

	if err := sendTelemetry(TapExecution, argsMap); err != nil {
		logger.Log.Debug(err)
		return
	}

	logger.Log.Debug("successfully reported telemetry of tap command")
}

func getExecutionTime(start time.Time) time.Duration {
	return time.Since(start)
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

func sendTelemetry(telemetryType telemetryType, argsMap map[string]interface{}) error {
	argsMap["telemetryType"] = telemetryType.String()
	argsMap["component"] = "mizu_cli"
	argsMap["buildTimestamp"] = mizu.BuildTimestamp
	argsMap["branch"] = mizu.Branch
	argsMap["version"] = mizu.SemVer
	argsMap["platform"] = mizu.Platform

	if machineId, err := machineid.ProtectedID("mizu"); err == nil {
		argsMap["machineId"] = machineId
	}

	jsonValue, _ := json.Marshal(argsMap)

	if resp, err := http.Post(telemetryUrl, "application/json", bytes.NewBuffer(jsonValue)); err != nil {
		return fmt.Errorf("ERROR: failed sending telemetry, err: %v, response %v", err, resp)
	}

	return nil
}
