package cmd

import (
	"github.com/spf13/cobra"
)

var pcapCmd = &cobra.Command{
	Use:   "pcap",
	Short: "Capture from a PCAP file using your Docker Daemon instead of Kubernetes.",
	RunE: func(cmd *cobra.Command, args []string) error {
		pcap()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pcapCmd)
}
