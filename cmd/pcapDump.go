package cmd

import (
	"errors"
	"path/filepath"
	"time"

	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// pcapDumpCmd represents the consolidated pcapdump command
var pcapDumpCmd = &cobra.Command{
	Use:   "pcapdump",
	Short: "Store all captured traffic (including decrypted TLS) in a PCAP file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Retrieve the kubeconfig path from the flag
		kubeconfig, _ := cmd.Flags().GetString(configStructs.PcapKubeconfig)

		// If kubeconfig is not provided, use the default location
		if kubeconfig == "" {
			if home := homedir.HomeDir(); home != "" {
				kubeconfig = filepath.Join(home, ".kube", "config")
			} else {
				return errors.New("kubeconfig flag not provided and no home directory available for default config location")
			}
		}

		debugEnabled, _ := cmd.Flags().GetBool("debug")
		if debugEnabled {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
			log.Debug().Msg("Debug logging enabled")
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}

		// Use the current context in kubeconfig
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Error().Err(err).Msg("Error building kubeconfig")
			return err
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			log.Error().Err(err).Msg("Error creating Kubernetes client")
			return err
		}

		// Parse the `--time` flag
		timeIntervalStr, _ := cmd.Flags().GetString("time")
		var cutoffTime *time.Time // Use a pointer to distinguish between provided and not provided
		if timeIntervalStr != "" {
			duration, err := time.ParseDuration(timeIntervalStr)
			if err != nil {
				log.Error().Err(err).Msg("Invalid time interval")
				return err
			}
			tempCutoffTime := time.Now().Add(-duration)
			cutoffTime = &tempCutoffTime
		}

		// Handle copy operation if the copy string is provided
		destDir, _ := cmd.Flags().GetString(configStructs.PcapDest)
		log.Info().Msg("Copying PCAP files")
		err = copyPcapFiles(clientset, config, destDir, cutoffTime)
		if err != nil {
			log.Error().Err(err).Msg("Error copying PCAP files")
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pcapDumpCmd)

	defaultPcapDumpConfig := configStructs.PcapDumpConfig{}
	if err := defaults.Set(&defaultPcapDumpConfig); err != nil {
		log.Debug().Err(err).Send()
	}

	pcapDumpCmd.Flags().String(configStructs.PcapTime, "", "Time interval (e.g., 10m, 1h) in the past for which the pcaps are copied")
	pcapDumpCmd.Flags().String(configStructs.PcapDest, "", "Local destination path for copied PCAP files (can not be used together with --enabled)")
	pcapDumpCmd.Flags().String(configStructs.PcapKubeconfig, "", "Path for kubeconfig (if not provided the default location will be checked)")
	pcapDumpCmd.Flags().Bool("debug", false, "Enable debug logging")
}
