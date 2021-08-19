package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/logger"
	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/cli/mizu/fsUtils"
	"github.com/up9inc/mizu/cli/mizu/version"
	"github.com/up9inc/mizu/cli/uiUtils"
)

var rootCmd = &cobra.Command{
	Use:   "mizu",
	Short: "A web traffic viewer for kubernetes",
	Long: `A web traffic viewer for kubernetes
Further info is available at https://github.com/up9inc/mizu`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := config.InitConfig(cmd); err != nil {
			logger.Log.Fatal(err)
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringSlice(config.SetCommandName, []string{}, fmt.Sprintf("Override values using --%s", config.SetCommandName))
}

func printNewVersionIfNeeded(versionChan chan string) {
	msg1 := <-versionChan
	if msg1 != "" {
		logger.Log.Infof(uiUtils.Yellow, msg1)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the tapCmd.
func Execute() {
	if err := fsUtils.EnsureDir(mizu.GetMizuFolderPath()); err != nil {
		logger.Log.Errorf("Failed to use mizu folder, %v", err)
	}
	logger.InitLogger()
	versionChan := make(chan string)
	defer printNewVersionIfNeeded(versionChan)
	go version.CheckNewerVersion(versionChan)
	cobra.CheckErr(rootCmd.Execute())
}
