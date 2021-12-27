package utils

import (
	"context"
	"github.com/up9inc/mizu/shared/logger"
	"os"
	"os/signal"
	"syscall"
)

func WaitForFinish(ctx context.Context, cancel context.CancelFunc) {
	logger.Log.Debugf("waiting for finish...")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// block until ctx cancel is called or termination signal is received
	select {
	case <-ctx.Done():
		logger.Log.Debugf("ctx done")
		break
	case <-sigChan:
		logger.Log.Debugf("Got termination signal, canceling execution...")
		cancel()
	}
}
