package cmd

import (
	"context"
	"path/filepath"

	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
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

		// Iterate over each pod to start the PCAP capture by updating the configuration in Kubernetes
		for _, pod := range workerPods.Items {
			err := setPcapConfigInKubernetes(clientset, pod.Name, namespace, "true", timeInterval, maxTime, maxSize)
			if err != nil {
				log.Error().Err(err).Msgf("Error setting PCAP config for pod %s", pod.Name)
				continue
			}
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
	pcapStartCmd.Flags().String("time-interval", defaultTapConfig.Misc.PcapTimeInterval, "Time interval for PCAP file rotation (e.g., 1m, 2h)")
	pcapStartCmd.Flags().String("max-time", defaultTapConfig.Misc.PcapMaxTime, "Maximum time for retaining old PCAP files (e.g., 24h)")
	pcapStartCmd.Flags().String("max-size", defaultTapConfig.Misc.PcapMaxSize, "Maximum size of PCAP files before deletion (e.g., 500MB, 10GB)")
}
