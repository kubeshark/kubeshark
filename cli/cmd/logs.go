package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/kubernetes"
	"github.com/up9inc/mizu/cli/logsUtils"
	"github.com/up9inc/mizu/cli/mizu"
	"os"
	"path"
)

var filePath string

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Create a zip file with logs for Github issue or troubleshoot",
	RunE: func(cmd *cobra.Command, args []string) error {
		kubernetesProvider, err := kubernetes.NewProvider(mizu.Config.View.KubeConfigPath)
		if err != nil {
			return nil
		}
		ctx, _ := context.WithCancel(context.Background())

		if filePath == "" {
			pwd, err := os.Getwd()
			if err != nil {
				mizu.Log.Errorf("Failed to get PWD, %v (try using `mizu logs -f <full path dest zip file>)`", err)
				return nil
			}
			filePath = path.Join(pwd, "mizu_logs.zip")
		}
		mizu.Log.Debugf("Using file path %s", filePath)

		if err := logsUtils.DumpLogs(kubernetesProvider, ctx, filePath); err != nil {
			mizu.Log.Errorf("Failed dump logs %v", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path for zip file (default current <pwd>\\mizu_logs.zip)")
}
