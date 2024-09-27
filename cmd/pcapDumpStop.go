package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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
	fmt.Println("Stopping PCAP capture.")

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
		err := writeStopConfigToFileInPod(context.Background(), clientset, config, pod.Name, namespace)
		if err != nil {
			log.Error().Err(err).Msgf("Error updating config file for pod %s", pod.Name)
			continue
		}
		fmt.Printf("PCAP capture stopped for pod %s\n", pod.Name)
	}

	fmt.Println("PCAP capture stopped successfully.")
	return nil
}

// writeStopConfigToFileInPod reads the existing config, updates the PCAP_DUMP_ENABLE value, and writes it back to the file
func writeStopConfigToFileInPod(ctx context.Context, clientset *kubernetes.Clientset, config *rest.Config, podName, namespace string) error {
	existingConfig, err := readConfigFileFromPod(ctx, clientset, config, podName, namespace, configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file from pod %s: %w", podName, err)
	}

	existingConfig["PCAP_DUMP_ENABLE"] = "false"

	err = writeConfigFileToPod(ctx, clientset, config, podName, namespace, configPath, existingConfig)
	if err != nil {
		return fmt.Errorf("failed to write config file to pod %s: %w", podName, err)
	}

	return nil
}
