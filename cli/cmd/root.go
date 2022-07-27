package cmd

import (
	"fmt"
	"time"

	"github.com/creasty/defaults"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/cli/mizu/fsUtils"
	"github.com/up9inc/mizu/cli/mizu/version"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/logger"
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
	defaultConfig := config.ConfigStruct{}
	if err := defaults.Set(&defaultConfig); err != nil {
		logger.Log.Debug(err)
	}

	rootCmd.PersistentFlags().StringSlice(config.SetCommandName, []string{}, fmt.Sprintf("Override values using --%s", config.SetCommandName))
	rootCmd.PersistentFlags().String(config.ConfigFilePathCommandName, defaultConfig.ConfigFilePath, fmt.Sprintf("Override config file path using --%s", config.ConfigFilePathCommandName))
}

func printNewVersionIfNeeded(versionChan chan string) {
	select {
	case versionMsg := <-versionChan:
		if versionMsg != "" {
			logger.Log.Infof(uiUtils.Yellow, versionMsg)
		}
	case <-time.After(2 * time.Second):
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the tapCmd.
func Execute() {
	if err := fsUtils.EnsureDir(mizu.GetMizuFolderPath()); err != nil {
		logger.Log.Errorf("Failed to use mizu folder, %v", err)
	}
	logger.InitLogger(fsUtils.GetLogFilePath())

	versionChan := make(chan string)
	defer printNewVersionIfNeeded(versionChan)
	go version.CheckNewerVersion(versionChan)

	cobra.CheckErr(rootCmd.Execute())
}
