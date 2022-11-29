package cmd

import (
	"strconv"
	"time"

	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/kubeshark"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version info",
	RunE: func(cmd *cobra.Command, args []string) error {
		timeStampInt, _ := strconv.ParseInt(kubeshark.BuildTimestamp, 10, 0)
		log.Info().
			Str("version", kubeshark.Ver).
			Str("branch", kubeshark.Branch).
			Str("commit-hash", kubeshark.GitCommitHash).
			Time("build-time", time.Unix(timeStampInt, 0))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	defaultVersionConfig := configStructs.VersionConfig{}
	if err := defaults.Set(&defaultVersionConfig); err != nil {
		log.Print(err)
	}

	versionCmd.Flags().BoolP(configStructs.DebugInfoVersionName, "d", defaultVersionConfig.DebugInfo, "Provide all information about version")

}
