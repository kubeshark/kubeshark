package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/misc"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version info",
	RunE: func(cmd *cobra.Command, args []string) error {
		timeStampInt, _ := strconv.ParseInt(misc.BuildTimestamp, 10, 0)
		if config.DebugMode {
			log.Info().
				Str("version", misc.Ver).
				Str("branch", misc.Branch).
				Str("commit-hash", misc.GitCommitHash).
				Time("build-time", time.Unix(timeStampInt, 0)).
				Send()
		} else {
			fmt.Println(misc.Ver)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
