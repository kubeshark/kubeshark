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
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Caller().Logger()
	goUtils.HandleExcWrapper(cmd.Execute)
}
