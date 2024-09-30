package cmd

import (
	"context"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// pcapstopCmd represents the pcapstop command
var pcapStopCmd = &cobra.Command{
	Use:   "pcapstop",
	Short: "Stop capturing traffic and close the PCAP dump",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Call the function to stop PCAP capture
		return stopPcap(cmd)
	},
}

func init() {
	rootCmd.AddCommand(pcapStopCmd)
}

func stopPcap(cmd *cobra.Command) error {
	// Load Kubernetes configuration
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

	// Get the list of worker pods
	workerPods, err := listWorkerPods(context.Background(), clientset, namespace)
	if err != nil {
		log.Error().Err(err).Msg("Error listing worker pods")
		return err
	}

	// Iterate over the worker pods and set config to stop pcap
	for _, pod := range workerPods.Items {
		err := setPcapConfigInKubernetes(clientset, pod.Name, namespace, "false", "", "", "")
		if err != nil {
			log.Error().Err(err).Msgf("Error setting PCAP config for pod %s", pod.Name)
			continue
		}
	}

	return nil
}
