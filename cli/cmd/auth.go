package cmd

import (
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate to up9 application",
}

func init() {
	rootCmd.AddCommand(authCmd)
}

