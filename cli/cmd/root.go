package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/mizu"
)

var rootCmd = &cobra.Command{
	Use:   "mizu",
	Short: "A web traffic viewer for kubernetes",
	Long: `A web traffic viewer for kubernetes
Further info is available at https://github.com/up9inc/mizu`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := mizu.MergeAllSettings(); err != nil {
			mizu.Log.Errorf("Invalid config, Exit %s", err)
			return errors.New(fmt.Sprintf("%v", err))
		}
		prettifiedConfig := mizu.GetConfigStr()
		mizu.Log.Debugf("Final Config: %s", prettifiedConfig)
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringSliceVar(&mizu.CommandLineValues, "set", []string{}, "Override values using --set")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the tapCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
