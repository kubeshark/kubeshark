package cmd

import (
	"log"
	"strconv"
	"time"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/config/configStructs"

	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/kubeshark"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version info",
	RunE: func(cmd *cobra.Command, args []string) error {
		if config.Config.Version.DebugInfo {
			timeStampInt, _ := strconv.ParseInt(kubeshark.BuildTimestamp, 10, 0)
			log.Printf("Version: %s \nBranch: %s (%s)", kubeshark.Ver, kubeshark.Branch, kubeshark.GitCommitHash)
			log.Printf("Build Time: %s (%s)", kubeshark.BuildTimestamp, time.Unix(timeStampInt, 0))

		} else {
			log.Printf("Version: %s (%s)", kubeshark.Ver, kubeshark.Branch)
		}
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
