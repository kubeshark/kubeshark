package utils

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func WaitForFinish(ctx context.Context, cancel context.CancelFunc) {
	log.Printf("waiting for finish...")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// block until ctx cancel is called or termination signal is received
	select {
	case <-ctx.Done():
		log.Printf("ctx done")
		break
	case <-sigChan:
		log.Printf("Got termination signal, canceling execution...")
		cancel()
	}
}
