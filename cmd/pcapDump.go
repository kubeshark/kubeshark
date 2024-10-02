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
	Short: "Manage PCAP operations: start, stop, or copy PCAP files",
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

		// Check flags for start, stop, copy
		start, _ := cmd.Flags().GetBool(configStructs.PcapStart)
		stop, _ := cmd.Flags().GetBool(configStructs.PcapStop)
		copy, _ := cmd.Flags().GetBool(configStructs.PcapCopy)

		// At least one of --start, --stop, or --copy must be provided
		if !start && !stop && !copy {
			return errors.New("at least one of --start, --stop, or --copy must be specified")
		}

		// Handle start operation if the start string is provided
		if start {
			timeInterval, _ := cmd.Flags().GetString(configStructs.PcapTimeInterval)
			maxTime, _ := cmd.Flags().GetString(configStructs.PcapMaxTime)
			maxSize, _ := cmd.Flags().GetString(configStructs.PcapMaxSize)
			err = startPcap(clientset, timeInterval, maxTime, maxSize)
			if err != nil {
				return err
			}
		}

		// Handle stop operation if the stop string is provided
		if stop {
			err = stopPcap(clientset)
			if err != nil {
				return err
			}
		}

		// Handle copy operation if the copy string is provided
		if copy {
			destDir, _ := cmd.Flags().GetString(configStructs.PcapDest)
			if destDir == "" {
				return errors.New("the --dest flag must be specified with --copy")
			}
			err = copyPcapFiles(clientset, config, destDir)
			if err != nil {
				return err
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

	pcapDumpCmd.Flags().Bool(configStructs.PcapStart, defaultPcapDumpConfig.PcapStart, "Start PCAP capture with the given name or parameters")
	pcapDumpCmd.Flags().Bool(configStructs.PcapStop, defaultPcapDumpConfig.PcapStop, "Stop PCAP capture with the given name or parameters")
	pcapDumpCmd.Flags().Bool(configStructs.PcapCopy, defaultPcapDumpConfig.PcapCopy, "Copy PCAP files with the given name or parameters")
	pcapDumpCmd.Flags().String(configStructs.PcapTimeInterval, defaultPcapDumpConfig.PcapTimeInterval, "Time interval for PCAP file rotation (used with --start)")
	pcapDumpCmd.Flags().String(configStructs.PcapMaxTime, defaultPcapDumpConfig.PcapMaxTime, "Maximum time for retaining old PCAP files (used with --start)")
	pcapDumpCmd.Flags().String(configStructs.PcapMaxSize, defaultPcapDumpConfig.PcapMaxSize, "Maximum size of PCAP files before deletion (used with --start)")
	pcapDumpCmd.Flags().String(configStructs.PcapDest, defaultPcapDumpConfig.PcapDest, "Local destination path for copied PCAP files (used with --copy)")
	pcapDumpCmd.Flags().String(configStructs.PcapKubeconfig, defaultPcapDumpConfig.PcapKubeconfig, "Absolute path to the kubeconfig file (optional)")
}
