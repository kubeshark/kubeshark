package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// pcapCmd represents the pcapstart command
var pcapStartCmd = &cobra.Command{
	Use:   "pcapstart",
	Short: "Start capturing traffic and generate a PCAP file",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create Kubernetes client
		kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
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

		// List worker pods
		workerPods, err := listWorkerPods(context.Background(), clientset, namespace)
		if err != nil {
			log.Error().Err(err).Msg("Error listing worker pods")
			return err
		}

		// Read the flags for time interval, max time, and max size
		timeInterval, _ := cmd.Flags().GetString("time-interval")
		maxTime, _ := cmd.Flags().GetString("max-time")
		maxSize, _ := cmd.Flags().GetString("max-size")

		// Iterate over each pod to start the PCAP capture by updating the config file
		for _, pod := range workerPods.Items {
			err := writeStartConfigToFileInPod(context.Background(), clientset, config, pod.Name, namespace, timeInterval, maxTime, maxSize)
			if err != nil {
				log.Error().Err(err).Msgf("Error updating config file for pod %s", pod.Name)
				continue
			}
			fmt.Printf("PCAP capture started for pod %s\n", pod.Name)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pcapStartCmd)

	defaultTapConfig := configStructs.TapConfig{}
	if err := defaults.Set(&defaultTapConfig); err != nil {
		log.Debug().Err(err).Send()
	}

	// Use full flag name without shorthand
	pcapStartCmd.Flags().String("time-interval", defaultTapConfig.Misc.TimeInterval, "Time interval for PCAP file rotation (e.g., 1m, 2h)")
	pcapStartCmd.Flags().String("max-time", defaultTapConfig.Misc.MaxTime, "Maximum time for retaining old PCAP files (e.g., 24h)")
	pcapStartCmd.Flags().String("max-size", defaultTapConfig.Misc.MaxSize, "Maximum size of PCAP files before deletion (e.g., 500MB, 10GB)")
}

// writeStartConfigToFileInPod writes config to start pcap in the worker pods
func writeStartConfigToFileInPod(ctx context.Context, clientset *kubernetes.Clientset, config *rest.Config, podName, namespace, timeInterval, maxTime, maxSize string) error {
	existingConfig, err := readConfigFileFromPod(ctx, clientset, config, podName, namespace, configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file from pod %s: %w", podName, err)
	}

	existingConfig["PCAP_DUMP_ENABLE"] = "true"
	if timeInterval != "" {
		existingConfig["TIME_INTERVAL"] = timeInterval
	}
	if maxTime != "" {
		existingConfig["MAX_TIME"] = maxTime
	}
	if maxSize != "" {
		existingConfig["MAX_SIZE"] = maxSize
	}

	err = writeConfigFileToPod(ctx, clientset, config, podName, namespace, configPath, existingConfig)
	if err != nil {
		return fmt.Errorf("failed to write config file to pod %s: %w", podName, err)
	}

	return nil
}
