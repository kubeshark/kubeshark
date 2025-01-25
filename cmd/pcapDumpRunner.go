package cmd

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kubeshark/gopacket/pcapgo"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	clientk8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	label  = "app.kubeshark.co/app=worker"
	srcDir = "pcapdump"
)

// NamespaceFiles represents the namespace and the files found in that namespace.
type NamespaceFiles struct {
	Namespace string   // The namespace in which the files were found
	SrcDir    string   // The source directory from which the files were listed
	Files     []string // List of files found in the namespace
}

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
func listFilesInPodDir(ctx context.Context, clientset *clientk8s.Clientset, config *rest.Config, podName string, namespaces []string, cutoffTime *time.Time) ([]NamespaceFiles, error) {
	var namespaceFilesList []NamespaceFiles

	for _, namespace := range namespaces {
		// Attempt to get the pod in the current namespace
		pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			continue
		}

		nodeName := pod.Spec.NodeName
		srcFilePath := filepath.Join("data", nodeName, srcDir)

		cmd := []string{"ls", srcFilePath}
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
			log.Error().Err(err).Msgf("failed to initialize executor for pod %s in namespace %s", podName, namespace)
			continue
		}

		var stdoutBuf bytes.Buffer
		var stderrBuf bytes.Buffer

		// Execute the command to list files
		err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
			Stdout: &stdoutBuf,
			Stderr: &stderrBuf,
		})
		if err != nil {
			log.Error().Err(err).Msgf("error listing files in pod %s in namespace %s: %s", podName, namespace, stderrBuf.String())
			continue
		}

		// Split the output (file names) into a list
		files := strings.Split(strings.TrimSpace(stdoutBuf.String()), "\n")
		if len(files) == 0 {
			log.Info().Msgf("No files found in directory %s in pod %s", srcFilePath, podName)
			continue
		}

		var filteredFiles []string

		// Filter files based on cutoff time if provided
		for _, file := range files {
			if cutoffTime != nil {
				parts := strings.Split(file, "-")
				if len(parts) < 2 {
					log.Warn().Msgf("Skipping file with invalid format: %s", file)
					continue
				}

				timestampStr := parts[len(parts)-2] + parts[len(parts)-1][:6] // Extract YYYYMMDDHHMMSS
				fileTime, err := time.Parse("20060102150405", timestampStr)
				if err != nil {
					log.Warn().Err(err).Msgf("Skipping file with unparsable timestamp: %s", file)
					continue
				}

				if fileTime.Before(*cutoffTime) {
					continue
				}
			}
			// Add file to filtered list
			filteredFiles = append(filteredFiles, file)
		}

		if len(filteredFiles) > 0 {
			namespaceFilesList = append(namespaceFilesList, NamespaceFiles{
				Namespace: namespace,
				SrcDir:    srcDir,
				Files:     filteredFiles,
			})
		}
	}

	if len(namespaceFilesList) == 0 {
		return nil, fmt.Errorf("no files found in pod %s across the provided namespaces", podName)
	}

	return namespaceFilesList, nil
}

// copyFileFromPod copies a single file from a pod to a local destination
func copyFileFromPod(ctx context.Context, clientset *kubernetes.Clientset, config *rest.Config, podName, namespace, srcDir, srcFile, destFile string) error {
	// Get the pod to retrieve its node name
	pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get pod %s in namespace %s: %w", podName, namespace, err)
	}

	// Construct the complete path using /data, the node name, srcDir, and srcFile
	nodeName := pod.Spec.NodeName
	srcFilePath := filepath.Join("data", nodeName, srcDir, srcFile)

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
		return fmt.Errorf("failed to initialize executor for pod %s in namespace %s: %w", podName, namespace, err)
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
		return fmt.Errorf("error copying file from pod %s in namespace %s: %s", podName, namespace, stderrBuf.String())
	}

	return nil
}

func mergePCAPs(outputFile string, inputFiles []string) error {
	// Create the output file
	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	bufWriter := bufio.NewWriter(f)
	defer bufWriter.Flush()

	// Create the PCAP writer
	writer := pcapgo.NewWriter(bufWriter)
	err = writer.WriteFileHeader(65536, 1)
	if err != nil {
		return fmt.Errorf("failed to write PCAP file header: %w", err)
	}

	for _, inputFile := range inputFiles {
		// Open the input file
		file, err := os.Open(inputFile)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to open %v", inputFile)
			continue
		}
		defer file.Close()

		// Create the PCAP reader for the input file
		reader, err := pcapgo.NewReader(file)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to create pcapng reader for %v", file.Name())
			continue
		}

		for {
			// Read packet data
			data, ci, err := reader.ReadPacketData()
			if err != nil {
				if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
					break
				}
				log.Error().Err(err).Msgf("Error reading packet from file %s", inputFile)
				break
			}

			// Write the packet to the output file
			err = writer.WritePacket(ci, data)
			if err != nil {
				log.Error().Err(err).Msgf("Error writing packet to output file")
				break
			}
		}
	}

	return nil
}

// copyPcapFiles function for copying the PCAP files from the worker pods
func copyPcapFiles(clientset *kubernetes.Clientset, config *rest.Config, destDir string, cutoffTime *time.Time) error {
	namespaceList, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Error listing namespaces")
		return err
	}

	var targetNamespaces []string
	for _, ns := range namespaceList.Items {
		targetNamespaces = append(targetNamespaces, ns.Name)
	}

	// List worker pods
	workerPods, err := listWorkerPods(context.Background(), clientset, targetNamespaces)
	if err != nil {
		log.Warn().Err(err).Msg("Error listing worker pods")
		return err
	}
	var currentFiles []string

	// Iterate over each pod to get the PCAP directory from config and copy files
	for _, pod := range workerPods {
		// Get the list of NamespaceFiles (files per namespace) and their source directories
		namespaceFiles, err := listFilesInPodDir(context.Background(), clientset, config, pod.Name, targetNamespaces, cutoffTime)
		if err != nil {
			log.Warn().Err(err).Send()
			continue
		}

		// Copy each file from the pod to the local destination for each namespace
		for _, nsFiles := range namespaceFiles {
			for _, file := range nsFiles.Files {
				destFile := filepath.Join(destDir, file)

				// Pass the correct namespace and related details to the function
				err = copyFileFromPod(context.Background(), clientset, config, pod.Name, nsFiles.Namespace, nsFiles.SrcDir, file, destFile)
				if err != nil {
					log.Error().Err(err).Msgf("Error copying file from pod %s in namespace %s", pod.Name, nsFiles.Namespace)
				} else {
					log.Info().Msgf("Copied %s from %s to %s", file, pod.Name, destFile)
				}

				currentFiles = append(currentFiles, destFile)
			}
		}
	}

	if len(currentFiles) == 0 {
		log.Error().Msgf("No files to merge")
		return nil
		// continue
	}

	// Generate a temporary filename based on the first file
	tempMergedFile := currentFiles[0] + "_temp"

	// Merge the PCAPs into the temporary file
	err = mergePCAPs(tempMergedFile, currentFiles)
	if err != nil {
		log.Error().Err(err).Msgf("Error merging files")
		return err
		// continue
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
		// continue
		return err
	}

	log.Info().Msgf("Merged file created: %s", finalMergedFile)

	return nil
}
