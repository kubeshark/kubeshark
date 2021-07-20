package mizu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/romana/rlog"
	"net/http"
)

const telemetryUrl = "https://us-east4-up9-prod.cloudfunctions.net/mizu-telemetry"

func ReportRun(cmd string, args interface{}) {
	if Branch != "main" {
		rlog.Debugf("reporting only on main branch")
		return
	}

	argsBytes, _ := json.Marshal(args)
	argsMap := map[string]string{"telemetry_type": "mizu_execution", "cmd": cmd, "args": string(argsBytes), "component": "mizu_cli"}
	argsMap["message"] = fmt.Sprintf("mizu %v - %v", argsMap["cmd"], string(argsBytes))

	jsonValue, _ := json.Marshal(argsMap)

	if resp, err := http.Post(telemetryUrl, "application/json", bytes.NewBuffer(jsonValue)); err != nil {
		rlog.Debugf("error sending telemtry err: %v, response %v", err, resp)
	} else {
		rlog.Debugf("Successfully reported telemetry")
	}
}
