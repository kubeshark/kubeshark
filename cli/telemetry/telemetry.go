package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/logger"
	"github.com/up9inc/mizu/cli/mizu"
	"net/http"
)

const telemetryUrl = "https://us-east4-up9-prod.cloudfunctions.net/mizu-telemetry"

func ReportRun(cmd string, args interface{}) {
	if !config.Config.Telemetry {
		logger.Log.Debugf("not reporting due to config value")
		return
	}

	if mizu.Branch != "main" && mizu.Branch != "develop" {
		logger.Log.Debugf("not reporting telemetry on private branches")
	}

	argsBytes, _ := json.Marshal(args)
	argsMap := map[string]string{
		"telemetry_type": "execution",
		"cmd":            cmd,
		"args":           string(argsBytes),
		"component":      "mizu_cli",
		"BuildTimestamp": mizu.BuildTimestamp,
		"Branch":         mizu.Branch,
		"version":        mizu.SemVer}
	argsMap["message"] = fmt.Sprintf("mizu %v - %v", argsMap["cmd"], string(argsBytes))

	jsonValue, _ := json.Marshal(argsMap)

	if resp, err := http.Post(telemetryUrl, "application/json", bytes.NewBuffer(jsonValue)); err != nil {
		logger.Log.Debugf("error sending telemetry err: %v, response %v", err, resp)
	} else {
		logger.Log.Debugf("Successfully reported telemetry")
	}
}
