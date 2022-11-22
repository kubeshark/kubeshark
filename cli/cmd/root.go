package cmd

import (
	"fmt"
	"time"

	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/cli/config"
	"github.com/kubeshark/kubeshark/cli/kubeshark"
	"github.com/kubeshark/kubeshark/cli/kubeshark/fsUtils"
	"github.com/kubeshark/kubeshark/cli/kubeshark/version"
	"github.com/kubeshark/kubeshark/cli/uiUtils"
	"github.com/kubeshark/kubeshark/logger"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kubeshark",
	Short: "A web traffic viewer for kubernetes",
	Long: `A web traffic viewer for kubernetes
Further info is available at https://github.com/kubeshark/kubeshark`,
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
	if err := fsUtils.EnsureDir(kubeshark.GetKubesharkFolderPath()); err != nil {
		logger.Log.Errorf("Failed to use kubeshark folder, %v", err)
	}
	logger.InitLogger(fsUtils.GetLogFilePath())

	versionChan := make(chan string)
	defer printNewVersionIfNeeded(versionChan)
	go version.CheckNewerVersion(versionChan)

	cobra.CheckErr(rootCmd.Execute())
}
