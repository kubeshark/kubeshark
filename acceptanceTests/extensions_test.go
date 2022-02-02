package acceptanceTests

import (
	"context"
	"github.com/go-redis/redis/v8"
	"os/exec"
	"testing"
	"time"
)

func TestRedis(t *testing.T) {
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

	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	for i := 0; i < defaultEntriesCount/5; i++ {
		requestErr := rdb.Ping(ctx).Err()
		if requestErr != nil {
			t.Errorf("failed to send redis request, err: %v", requestErr)
			return
		}
	}

	for i := 0; i < defaultEntriesCount/5; i++ {
		requestErr := rdb.Set(ctx, "key", "value", -1).Err()
		if requestErr != nil {
			t.Errorf("failed to send redis request, err: %v", requestErr)
			return
		}
	}

	for i := 0; i < defaultEntriesCount/5; i++ {
		requestErr := rdb.Exists(ctx, "key").Err()
		if requestErr != nil {
			t.Errorf("failed to send redis request, err: %v", requestErr)
			return
		}
	}

	for i := 0; i < defaultEntriesCount/5; i++ {
		requestErr := rdb.Get(ctx, "key").Err()
		if requestErr != nil {
			t.Errorf("failed to send redis request, err: %v", requestErr)
			return
		}
	}

	for i := 0; i < defaultEntriesCount/5; i++ {
		requestErr := rdb.Del(ctx, "key").Err()
		if requestErr != nil {
			t.Errorf("failed to send redis request, err: %v", requestErr)
			return
		}
	}

	time.Sleep(1 * time.Hour)
}
