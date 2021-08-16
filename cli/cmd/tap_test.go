package cmd_test

import (
	"encoding/json"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/cmd"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/logger"
	"io/ioutil"
	"net/http"
	"syscall"
	"testing"
	"time"
)

func TestIntegrationTap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	commandMock := cobra.Command{}
	if err := config.InitConfig(&commandMock); err != nil {
		t.Errorf("error ")
	}

	config.Config.AgentImage = "ci:latest"
	config.Config.Tap.Namespaces = []string{"mizu-tests"}
	config.Config.Telemetry = false
	go cmd.RunMizuTap()

	time.Sleep(30 * time.Second)

	for i := 0; i < 100; i++ {
		http.Get("http://localhost:8080/api/v1/namespaces/mizu-tests/services/httpbin/proxy/get")
	}

	time.Sleep(5 * time.Second)

	resp, _ := http.Get("http://localhost:8899/mizu/api/generalStats")
	data, _ := ioutil.ReadAll(resp.Body)

	var generalStats map[string]interface{}
	json.Unmarshal(data, &generalStats)

	entriesCount := generalStats["EntriesCount"]

	logger.Log.Infof("%v", entriesCount)

	if entriesCount != 100.0 {
		t.Errorf("test failed")
	}

	syscall.Kill(syscall.Getpid(), syscall.SIGINT)

	time.Sleep(5 * time.Second)
}
