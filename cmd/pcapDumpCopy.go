package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	namespace = "default"
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
			srcDir, err := getWorkerSource(clientset, pod.Name, pod.Namespace)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to read the worker source dir for %s", pod.Name)
			}

			// List files in the PCAP directory on the pod
			files, err := listFilesInPodDir(context.Background(), clientset, config, pod.Name, namespace, srcDir)
			if err != nil {
				log.Error().Err(err).Msgf("Error listing files in pod %s", pod.Name)
				continue
			}

			var currentFiles []string

			// Copy each file from the pod to the local destination
			for _, file := range files {
				srcFile := filepath.Join(srcDir, file)
				destFile := filepath.Join(destDir, pod.Name+"_"+file)

				err = copyFileFromPod(context.Background(), clientset, config, pod.Name, namespace, srcFile, destFile)
				if err != nil {
					log.Error().Err(err).Msgf("Error copying file from pod %s", pod.Name)
				}

				currentFiles = append(currentFiles, destFile)
			}
			if len(currentFiles) == 0 {
				log.Error().Msgf("No files to merge for pod %s", pod.Name)
				continue
			}

			// Generate a temporary filename based on the first file
			tempMergedFile := currentFiles[0] + "_temp"

			// Merge the PCAPs into the temporary file
			err = mergePCAPs(tempMergedFile, currentFiles)
			if err != nil {
				log.Error().Err(err).Msgf("Error merging file from pod %s", pod.Name)
				continue
			}

			// Remove the original files after merging
			for _, file := range currentFiles {
				err := os.Remove(file)
				if err != nil {
					log.Error().Err(err).Msgf("Error removing file %s", file)
				}
			}

			// Rename the temp file to the final name (removing "_temp")
			finalMergedFile := strings.TrimSuffix(tempMergedFile, "_temp")
			err = os.Rename(tempMergedFile, finalMergedFile)
			if err != nil {
				log.Error().Err(err).Msgf("Error renaming merged file %s", tempMergedFile)
				continue
			}

			log.Info().Msgf("Merged file created: %s", finalMergedFile)
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
