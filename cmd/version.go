package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/kubeshark"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version info",
	RunE: func(cmd *cobra.Command, args []string) error {
		timeStampInt, _ := strconv.ParseInt(kubeshark.BuildTimestamp, 10, 0)
		if config.DebugMode {
			log.Info().
				Str("version", kubeshark.Ver).
				Str("branch", kubeshark.Branch).
				Str("commit-hash", kubeshark.GitCommitHash).
				Time("build-time", time.Unix(timeStampInt, 0)).
				Send()
		} else {
			fmt.Println(kubeshark.Ver)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
