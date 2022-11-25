package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/kubeshark/version"
	"github.com/kubeshark/kubeshark/uiUtils"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kubeshark",
	Short: "A web traffic viewer for kubernetes",
	Long: `A web traffic viewer for kubernetes
Further info is available at https://github.com/kubeshark/kubeshark`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := config.InitConfig(cmd); err != nil {
			log.Fatal(err)
		}

		return nil
	},
}

func init() {
	defaultConfig := config.CreateDefaultConfig()
	if err := defaults.Set(&defaultConfig); err != nil {
		log.Print(err)
	}

	rootCmd.PersistentFlags().StringSlice(config.SetCommandName, []string{}, fmt.Sprintf("Override values using --%s", config.SetCommandName))
	rootCmd.PersistentFlags().String(config.ConfigFilePathCommandName, defaultConfig.ConfigFilePath, fmt.Sprintf("Override config file path using --%s", config.ConfigFilePathCommandName))
}

func printNewVersionIfNeeded(versionChan chan string) {
	select {
	case versionMsg := <-versionChan:
		if versionMsg != "" {
			log.Printf(uiUtils.Yellow, versionMsg)
		}
	case <-time.After(2 * time.Second):
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the tapCmd.
func Execute() {
	versionChan := make(chan string)
	defer printNewVersionIfNeeded(versionChan)
	go version.CheckNewerVersion(versionChan)

	cobra.CheckErr(rootCmd.Execute())
}
