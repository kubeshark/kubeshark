package utils

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
)

func WaitForTermination(ctx context.Context, cancel context.CancelFunc) {
	log.Debug().Msg("Waiting to finish...")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// block until ctx cancel is called or termination signal is received
	select {
	case <-ctx.Done():
		log.Debug().Msg("Context done.")
		break
	case <-sigChan:
		log.Debug().Msg("Got a termination signal, canceling execution...")
		cancel()
	}
}
