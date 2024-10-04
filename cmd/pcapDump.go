package cmd

import (
	"errors"
	"path/filepath"

	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// pcapDumpCmd represents the consolidated pcapdump command
var pcapDumpCmd = &cobra.Command{
	Use:   "pcapdump",
	Short: "Manage PCAP dump operations: start, stop, or copy PCAP files",
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

		// Handle copy operation if the copy string is provided

		enabled, err := cmd.Flags().GetString(configStructs.PcapDumpEnable)
		if err != nil {
			log.Error().Err(err).Msg("Error getting pcapdump enable flag")
			return err
		}
		if enabled == "unknown" {
			destDir, _ := cmd.Flags().GetString(configStructs.PcapDest)
			log.Info().Msg("Copying PCAP files")
			err = copyPcapFiles(clientset, config, destDir)
			if err != nil {
				log.Error().Err(err).Msg("Error copying PCAP files")
				return err
			}
		} else {
			// Handle start operation if the start string is provided
			if enabled == "true" || enabled == "false" {
				log.Info().Msg("Starting/stopping PCAP dump")
				timeInterval, _ := cmd.Flags().GetString(configStructs.PcapTimeInterval)
				maxTime, _ := cmd.Flags().GetString(configStructs.PcapMaxTime)
				maxSize, _ := cmd.Flags().GetString(configStructs.PcapMaxSize)
				err = startStopPcap(clientset, enabled, timeInterval, maxTime, maxSize)
				if err != nil {
					log.Error().Err(err).Msg("Error starting/stopping PCAP dump")
					return err
				}

				if enabled == "true" {
					log.Info().Msg("Pcapdump started successfully")
					return nil
				}

				if enabled == "false" {
					log.Info().Msg("Pcapdump stopped successfully")
					return nil
				}
				return errors.New("Invalid value for pcapdump command")
			}

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

	pcapDumpCmd.Flags().String(configStructs.PcapTimeInterval, defaultPcapDumpConfig.PcapTimeInterval, "Time interval for PCAP file rotation (used with --start)")
	pcapDumpCmd.Flags().String(configStructs.PcapMaxTime, defaultPcapDumpConfig.PcapMaxTime, "Maximum time for retaining old PCAP files (used with --start)")
	pcapDumpCmd.Flags().String(configStructs.PcapMaxSize, defaultPcapDumpConfig.PcapMaxSize, "Maximum size of PCAP files before deletion (used with --start)")
	pcapDumpCmd.Flags().String(configStructs.PcapDest, defaultPcapDumpConfig.PcapDest, "Local destination path for copied PCAP files (can not be used together with --enabled)")
	pcapDumpCmd.Flags().String(configStructs.PcapKubeconfig, defaultPcapDumpConfig.PcapKubeconfig, "Absolute path to the kubeconfig file (optional)")
	pcapDumpCmd.Flags().String(configStructs.PcapDumpEnable, defaultPcapDumpConfig.PcapDumpEnable, "Enabled/Disable to pcap dumps (can not be used together with --dest)")

}
