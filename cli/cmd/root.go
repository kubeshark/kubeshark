package cmd

import (
	"errors"
	"fmt"
	"github.com/romana/rlog"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/uiUtils"
	"os"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "mizu",
	Short: "A web traffic viewer for kubernetes",
	Long: `A web traffic viewer for kubernetes
Further info is available at https://github.com/up9inc/mizu`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if verbose {
			_ = os.Setenv("RLOG_LOG_LEVEL", "DEBUG")
			rlog.UpdateEnv()
		}
		if err := config.MergeAllSettings(); err != nil {
			rlog.Infof("Invalid config, Exit %s", err)
			return errors.New(fmt.Sprintf("%v", err))
		}
		prettifiedConfig, _ := uiUtils.PrettyJson(config.GetConfig())
		rlog.Debugf("Final Config: %s", prettifiedConfig)
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringSliceVar(&config.CommandLineValues, "set", []string{}, "Override values using --set")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Verbose logging")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the tapCmd.
func Execute() {
	rlog.Debug("Executing Root command")
	cobra.CheckErr(rootCmd.Execute())
}
