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

const (
	configPath = "/app/config/pcap_config.txt"
	namespace  = "default"
)

// pcapCopyCmd represents the pcapcopy command
var pcapCopyCmd = &cobra.Command{
	Use:   "pcapcopy",
	Short: "Copy PCAP files from worker pods to the local destination",
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

		// Destination directory for the files
		destDir, _ := cmd.Flags().GetString(configStructs.PcapDumpCopy)

		// Iterate over each pod to get the PCAP directory from config and copy files
		for _, pod := range workerPods.Items {
			// Read the config file from the pod to get the PCAP_DIR value
			configMap, err := readConfigFileFromPod(context.Background(), clientset, config, pod.Name, namespace, configPath)
			if err != nil {
				log.Error().Err(err).Msgf("Error reading config file from pod %s", pod.Name)
				continue
			}

			// Use the PCAP_DIR value from the config file
			srcDir := configMap["PCAP_DIR"]
			if srcDir == "" {
				log.Error().Msgf("PCAP_DIR not found in config for pod %s", pod.Name)
				continue
			}

			// List files in the PCAP directory on the pod
			files, err := listFilesInPodDir(context.Background(), clientset, config, pod.Name, namespace, srcDir)
			if err != nil {
				log.Error().Err(err).Msgf("Error listing files in pod %s", pod.Name)
				continue
			}

			// Copy each file from the pod to the local destination
			for _, file := range files {
				srcFile := filepath.Join(srcDir, file)
				destFile := filepath.Join(destDir, pod.Name+"_"+file)

				err = copyFileFromPod(context.Background(), clientset, config, pod.Name, namespace, srcFile, destFile)
				if err != nil {
					log.Error().Err(err).Msgf("Error copying file from pod %s", pod.Name)
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pcapCopyCmd)

	defaultTapConfig := configStructs.TapConfig{}
	if err := defaults.Set(&defaultTapConfig); err != nil {
		log.Debug().Err(err).Send()
	}

	// Use full flag name without shorthand
	pcapCopyCmd.Flags().String(configStructs.PcapDumpCopy, defaultTapConfig.Misc.PcapDest, "Local destination path for the copied files (required)")
	_ = pcapCopyCmd.MarkFlagRequired("dest")
}
