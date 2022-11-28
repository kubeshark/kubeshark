package main

import (
	"os"
	"time"

	"github.com/kubeshark/kubeshark/cmd"
	"github.com/kubeshark/kubeshark/cmd/goUtils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	goUtils.HandleExcWrapper(cmd.Execute)
}
