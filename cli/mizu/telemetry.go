package mizu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const telemetryUrl = "https://us-east4-up9-prod.cloudfunctions.net/mizu-telemetry"

func ReportRun(cmd string, args interface{}) {
	if !Config.Telemetry {
		Log.Debugf("not reporting due to config value")
		return
	}

	if Branch != "main" && Branch != "develop" {
		Log.Debugf("not reporting telemetry on private branches")
	}

	argsBytes, _ := json.Marshal(args)
	argsMap := map[string]string{
		"telemetry_type": "execution",
		"cmd":            cmd,
		"args":           string(argsBytes),
		"component":      "mizu_cli",
		"BuildTimestamp": BuildTimestamp,
		"Branch":         Branch,
		"version":        SemVer}
	argsMap["message"] = fmt.Sprintf("mizu %v - %v", argsMap["cmd"], string(argsBytes))

	jsonValue, _ := json.Marshal(argsMap)

	if resp, err := http.Post(telemetryUrl, "application/json", bytes.NewBuffer(jsonValue)); err != nil {
		Log.Debugf("error sending telemetry err: %v, response %v", err, resp)
	} else {
		Log.Debugf("Successfully reported telemetry")
	}
}
