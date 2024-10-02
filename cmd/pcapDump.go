package cmd

import (
	"errors"
	"flag"
	"fmt"
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
		var kubeconfig *string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()

		// Use the current context in kubeconfig
		config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
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
		start, _ := cmd.Flags().GetString("start")
		stop, _ := cmd.Flags().GetString("stop")
		copy, _ := cmd.Flags().GetString("copy")

		// At least one of --start, --stop, or --copy must be provided
		if start == "" && stop == "" && copy == "" {
			return errors.New("at least one of --start, --stop, or --copy must be specified")
		}

		// Handle start operation if the start string is provided
		if start != "" {
			timeInterval, _ := cmd.Flags().GetString("time-interval")
			maxTime, _ := cmd.Flags().GetString("max-time")
			maxSize, _ := cmd.Flags().GetString("max-size")
			err = startPcap(clientset, timeInterval, maxTime, maxSize)
			if err != nil {
				return err
			}
			fmt.Printf("Started PCAP capture with parameters: %v, %v, %v\n", timeInterval, maxTime, maxSize)
		}

		// Handle stop operation if the stop string is provided
		if stop != "" {
			err = stopPcap(clientset)
			if err != nil {
				return err
			}
			fmt.Println("Stopped PCAP capture")
		}

		// Handle copy operation if the copy string is provided
		if copy != "" {
			destDir, _ := cmd.Flags().GetString("dest")
			if destDir == "" {
				return errors.New("the --dest flag must be specified with --copy")
			}
			fmt.Printf("Copying PCAP files to destination: %v\n", destDir)
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

	defaultTapConfig := configStructs.TapConfig{}
	if err := defaults.Set(&defaultTapConfig); err != nil {
		log.Debug().Err(err).Send()
	}

	pcapDumpCmd.Flags().String("start", "", "Start PCAP capture with the given name or parameters")
	pcapDumpCmd.Flags().String("stop", "", "Stop PCAP capture with the given name or parameters")
	pcapDumpCmd.Flags().String(configStructs.PcapCopy, defaultTapConfig.Misc.PcapCopy, "Copy PCAP files with the given name or parameters")
	pcapDumpCmd.Flags().String("time-interval", defaultTapConfig.Misc.PcapTimeInterval, "Time interval for PCAP file rotation (used with --start)")
	pcapDumpCmd.Flags().String("max-time", defaultTapConfig.Misc.PcapMaxTime, "Maximum time for retaining old PCAP files (used with --start)")
	pcapDumpCmd.Flags().String("max-size", defaultTapConfig.Misc.PcapMaxSize, "Maximum size of PCAP files before deletion (used with --start)")
	pcapDumpCmd.Flags().String("dest", defaultTapConfig.Misc.PcapDest, "Local destination path for copied PCAP files (used with --copy)")
	pcapDumpCmd.Flags().String("kubeconfig", "", "Absolute path to the kubeconfig file (optional)")
}
