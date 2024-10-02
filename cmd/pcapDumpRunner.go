package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kubeshark/gopacket"
	"github.com/kubeshark/gopacket/layers"
	"github.com/kubeshark/gopacket/pcapgo"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	clientk8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

const label = "app.kubeshark.co/app=worker"
const SELF_RESOURCES_PREFIX = "kubeshark-"
const SUFFIX_CONFIG_MAP = "config-map"

// listWorkerPods fetches all worker pods from multiple namespaces
func listWorkerPods(ctx context.Context, clientset *clientk8s.Clientset, namespaces []string) ([]corev1.Pod, error) {
	var allPods []corev1.Pod
	labelSelector := label

	for _, namespace := range namespaces {
		// List all pods matching the label in the current namespace
		pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list worker pods in namespace %s: %w", namespace, err)
		}

		// Accumulate the pods
		allPods = append(allPods, pods.Items...)
	}

	return allPods, nil
}

// listFilesInPodDir lists all files in the specified directory inside the pod across multiple namespaces
func listFilesInPodDir(ctx context.Context, clientset *clientk8s.Clientset, config *rest.Config, podName string, namespaces []string, configMapName, configMapKey string) (string, []string, error) {
	var allFiles []string
	var srcDir string
	var ok bool

	for _, namespace := range namespaces {
		// Step 1: Retrieve the ConfigMap for the current namespace
		configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
		if err != nil {
			return "", nil, fmt.Errorf("failed to retrieve ConfigMap %s in namespace %s: %w", configMapName, namespace, err)
		}

		// Step 2: Get the srcDir from the ConfigMap data
		srcDir, ok = configMap.Data[configMapKey]
		if !ok || srcDir == "" {
			return "", nil, fmt.Errorf("directory path not found in ConfigMap %s under key %s", configMapName, configMapKey)
		}

		// Step 3: List files in the retrieved srcDir
		cmd := []string{"ls", srcDir}
		req := clientset.CoreV1().RESTClient().Post().
			Resource("pods").
			Name(podName).
			Namespace(namespace).
			SubResource("exec").
			Param("container", "sniffer").
			Param("stdout", "true").
			Param("stderr", "true").
			Param("command", cmd[0]).
			Param("command", cmd[1])

		exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
		if err != nil {
			return "", nil, fmt.Errorf("failed to initialize executor for pod %s in namespace %s: %w", podName, namespace, err)
		}

		var stdoutBuf bytes.Buffer
		var stderrBuf bytes.Buffer

		// Stream the result of the ls command
		err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
			Stdout: &stdoutBuf,
			Stderr: &stderrBuf,
		})
		if err != nil {
			return "", nil, fmt.Errorf("error listing files in pod %s in namespace %s: %w. Stderr: %s", podName, namespace, err, stderrBuf.String())
		}

		// Split the output (file names) into a list and accumulate them
		files := strings.Split(strings.TrimSpace(stdoutBuf.String()), "\n")
		allFiles = append(allFiles, files...)
	}

	if len(allFiles) == 0 {
		return "", nil, fmt.Errorf("no files found in pod %s across the provided namespaces", podName)
	}

	return srcDir, allFiles, nil
}

// copyFileFromPod copies a single file from a pod to a local destination across multiple namespaces
func copyFileFromPod(ctx context.Context, clientset *clientk8s.Clientset, config *rest.Config, podName, srcDir, srcFile, destFile string, namespaces []string) error {
	for _, namespace := range namespaces {
		// Get the pod to retrieve its node name
		pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			log.Error().Err(err).Msgf("Failed to get pod %s in namespace %s", podName, namespace)
			continue // Move to the next namespace
		}

		// Construct the complete path using /data, the node name, srcDir, and srcFile
		nodeName := pod.Spec.NodeName
		srcFilePath := filepath.Join("/data", nodeName, srcDir, srcFile)

		log.Warn().Msgf("Copying file %v", srcFilePath)

		// Execute the `cat` command to read the file at the srcFilePath
		cmd := []string{"cat", srcFilePath}
		req := clientset.CoreV1().RESTClient().Post().
			Resource("pods").
			Name(podName).
			Namespace(namespace).
			SubResource("exec").
			Param("container", "sniffer").
			Param("stdout", "true").
			Param("stderr", "true").
			Param("command", cmd[0]).
			Param("command", cmd[1])

		exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to initialize executor for pod %s in namespace %s", podName, namespace)
			continue // Move to the next namespace
		}

		// Create the local file to write the content to
		outFile, err := os.Create(destFile)
		if err != nil {
			return fmt.Errorf("failed to create destination file: %w", err)
		}
		defer outFile.Close()

		// Capture stderr for error logging
		var stderrBuf bytes.Buffer

		// Stream the file content from the pod to the local file
		err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
			Stdout: outFile,
			Stderr: &stderrBuf,
		})
		if err != nil {
			log.Error().Err(err).Msgf("Error copying file from pod %s in namespace %s: %s", podName, namespace, stderrBuf.String())
			continue // Move to the next namespace
		}

		// If the copy is successful, break out of the loop as there's no need to continue
		log.Info().Msgf("Successfully copied file from pod %s in namespace %s to %s", podName, namespace, destFile)
		return nil
	}

	// If the file couldn't be copied from any namespace, return an error
	return fmt.Errorf("failed to copy file from pod %s across all provided namespaces", podName)
}

func mergePCAPs(outputFile string, inputFiles []string) error {
	// Create the output file
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	// Create a pcap writer for the output file
	writer := pcapgo.NewWriter(f)
	err = writer.WriteFileHeader(65536, layers.LinkTypeEthernet) // Snapshot length and LinkType
	if err != nil {
		return err
	}

	for _, inputFile := range inputFiles {
		// Open each input file
		file, err := os.Open(inputFile)
		if err != nil {
			return err
		}
		defer file.Close()

		reader, err := pcapgo.NewReader(file)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to create pcapng reader for %v", file.Name())
			return err
		}

		// Create the packet source
		packetSource := gopacket.NewPacketSource(reader, layers.LinkTypeEthernet)

		for packet := range packetSource.Packets() {
			err := writer.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// setPcapConfigInKubernetes sets the PCAP config for all pods across multiple namespaces
func setPcapConfigInKubernetes(ctx context.Context, clientset *clientk8s.Clientset, podName string, namespaces []string, enabledPcap, timeInterval, maxTime, maxSize string) error {
	for _, namespace := range namespaces {
		// Load the existing ConfigMap in the current namespace
		configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(ctx, "kubeshark-config-map", metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to load ConfigMap in namespace %s: %w", namespace, err)
		}

		// Update the values with user-provided input
		configMap.Data["PCAP_TIME_INTERVAL"] = timeInterval
		configMap.Data["PCAP_MAX_SIZE"] = maxSize
		configMap.Data["PCAP_MAX_TIME"] = maxTime
		configMap.Data["PCAP_DUMP_ENABLE"] = enabledPcap

		// Apply the updated ConfigMap back to the cluster in the current namespace
		_, err = clientset.CoreV1().ConfigMaps(namespace).Update(ctx, configMap, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update ConfigMap in namespace %s: %w", namespace, err)
		}
	}

	return nil
}

// startPcap function for starting the PCAP capture
func startPcap(clientset *kubernetes.Clientset, timeInterval, maxTime, maxSize string) error {
	kubernetesProvider, err := getKubernetesProviderForCli(false, false)
	if err != nil {
		log.Error().Err(err).Send()
		return err
	}

	targetNamespaces := kubernetesProvider.GetNamespaces()

	// List worker pods
	workerPods, err := listWorkerPods(context.Background(), clientset, targetNamespaces)
	if err != nil {
		log.Error().Err(err).Msg("Error listing worker pods")
		return err
	}

	// Iterate over each pod to start the PCAP capture by updating the configuration in Kubernetes
	for _, pod := range workerPods {
		err := setPcapConfigInKubernetes(context.Background(), clientset, pod.Name, targetNamespaces, "true", timeInterval, maxTime, maxSize)
		if err != nil {
			log.Error().Err(err).Msgf("Error setting PCAP config for pod %s", pod.Name)
			continue
		}
	}
	return nil
}

// stopPcap function for stopping the PCAP capture
func stopPcap(clientset *kubernetes.Clientset) error {
	kubernetesProvider, err := getKubernetesProviderForCli(false, false)
	if err != nil {
		log.Error().Err(err).Send()
		return err
	}

	targetNamespaces := kubernetesProvider.GetNamespaces()

	// Get the list of worker pods
	workerPods, err := listWorkerPods(context.Background(), clientset, targetNamespaces)
	if err != nil {
		log.Error().Err(err).Msg("Error listing worker pods")
		return err
	}

	// Iterate over the worker pods and set config to stop pcap
	for _, pod := range workerPods {
		err := setPcapConfigInKubernetes(context.Background(), clientset, pod.Name, targetNamespaces, "false", "", "", "")
		if err != nil {
			log.Error().Err(err).Msgf("Error setting PCAP config for pod %s", pod.Name)
			continue
		}
	}
	return nil
}

// copyPcapFiles function for copying the PCAP files from the worker pods
func copyPcapFiles(clientset *kubernetes.Clientset, config *rest.Config, destDir string) error {
	kubernetesProvider, err := getKubernetesProviderForCli(false, false)
	if err != nil {
		log.Error().Err(err).Send()
		return err
	}

	targetNamespaces := kubernetesProvider.GetNamespaces()

	// List worker pods
	workerPods, err := listWorkerPods(context.Background(), clientset, targetNamespaces)
	if err != nil {
		log.Error().Err(err).Msg("Error listing worker pods")
		return err
	}

	// Iterate over each pod to get the PCAP directory from config and copy files
	for _, pod := range workerPods {
		// List files in the PCAP directory on the pod
		srcDir, files, err := listFilesInPodDir(context.Background(), clientset, config, pod.Name, targetNamespaces, SELF_RESOURCES_PREFIX+SUFFIX_CONFIG_MAP, "PCAP_SRC_DIR")
		if err != nil {
			log.Error().Err(err).Msgf("Error listing files in pod %s", pod.Name)
			continue
		}

		var currentFiles []string

		// Copy each file from the pod to the local destination
		for _, file := range files {
			destFile := filepath.Join(destDir, pod.Name+"_"+file)

			log.Warn().Msgf("Got srcDir %v", srcDir)

			err = copyFileFromPod(context.Background(), clientset, config, pod.Name, srcDir, file, destFile, targetNamespaces)
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
}
