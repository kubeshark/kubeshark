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
	"sync"
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
	label                 = "app.kubeshark.co/app=worker"
	srcDir                = "pcapdump"
	maxSnaplen     uint32 = 262144
	maxTimePerFile        = time.Minute * 5
)

// PodFileInfo represents information about a pod, its namespace, and associated files
type PodFileInfo struct {
	Pod         corev1.Pod
	SrcDir      string
	Files       []string
	CopiedFiles []string
}

// listWorkerPods fetches all worker pods from multiple namespaces
func listWorkerPods(ctx context.Context, clientset *clientk8s.Clientset, namespaces []string) ([]*PodFileInfo, error) {
	var podFileInfos []*PodFileInfo
	var errs []error
	labelSelector := label

	for _, namespace := range namespaces {
		// List all pods matching the label in the current namespace
		pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to list worker pods in namespace %s: %w", namespace, err))
			continue
		}

		for _, pod := range pods.Items {
			podFileInfos = append(podFileInfos, &PodFileInfo{
				Pod: pod,
			})
		}
	}

	return podFileInfos, errors.Join(errs...)
}

// listFilesInPodDir lists all files in the specified directory inside the pod across multiple namespaces
func listFilesInPodDir(ctx context.Context, clientset *clientk8s.Clientset, config *rest.Config, pod *PodFileInfo, cutoffTime *time.Time) error {
	nodeName := pod.Pod.Spec.NodeName
	srcFilePath := filepath.Join("data", nodeName, srcDir)

	cmd := []string{"ls", srcFilePath}
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Pod.Name).
		Namespace(pod.Pod.Namespace).
		SubResource("exec").
		Param("container", "sniffer").
		Param("stdout", "true").
		Param("stderr", "true").
		Param("command", cmd[0]).
		Param("command", cmd[1])

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return err
	}

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	// Execute the command to list files
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdoutBuf,
		Stderr: &stderrBuf,
	})
	if err != nil {
		return err
	}

	// Split the output (file names) into a list
	files := strings.Split(strings.TrimSpace(stdoutBuf.String()), "\n")
	if len(files) == 0 {
		// No files were found in the target dir for this pod
		return nil
	}

	var filteredFiles []string
	var fileProcessingErrs []error
	// Filter files based on cutoff time if provided
	for _, file := range files {
		if cutoffTime != nil {
			parts := strings.Split(file, "-")
			if len(parts) < 2 {
				continue
			}

			timestampStr := parts[len(parts)-2] + parts[len(parts)-1][:6] // Extract YYYYMMDDHHMMSS
			fileTime, err := time.Parse("20060102150405", timestampStr)
			if err != nil {
				fileProcessingErrs = append(fileProcessingErrs, fmt.Errorf("failed parse file timestamp %s: %w", file, err))
				continue
			}

			if fileTime.Before(*cutoffTime) {
				continue
			}
		}
		// Add file to filtered list
		filteredFiles = append(filteredFiles, file)
	}

	pod.SrcDir = srcDir
	pod.Files = filteredFiles

	return errors.Join(fileProcessingErrs...)
}

// copyFileFromPod copies a single file from a pod to a local destination
func copyFileFromPod(ctx context.Context, clientset *kubernetes.Clientset, config *rest.Config, pod *PodFileInfo, srcFile, destFile string) error {
	// Construct the complete path using /data, the node name, srcDir, and srcFile
	nodeName := pod.Pod.Spec.NodeName
	srcFilePath := filepath.Join("data", nodeName, srcDir, srcFile)

	// Execute the `cat` command to read the file at the srcFilePath
	cmd := []string{"cat", srcFilePath}
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Pod.Name).
		Namespace(pod.Pod.Namespace).
		SubResource("exec").
		Param("container", "sniffer").
		Param("stdout", "true").
		Param("stderr", "true").
		Param("command", cmd[0]).
		Param("command", cmd[1])

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("failed to initialize executor for pod %s in namespace %s: %w", pod.Pod.Name, pod.Pod.Namespace, err)
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
		return err
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

	bufWriter := bufio.NewWriterSize(f, 4*1024*1024)
	defer bufWriter.Flush()

	// Create the PCAP writer
	writer := pcapgo.NewWriter(bufWriter)
	err = writer.WriteFileHeader(maxSnaplen, 1)
	if err != nil {
		return fmt.Errorf("failed to write PCAP file header: %w", err)
	}

	var mergingErrs []error

	for _, inputFile := range inputFiles {
		// Open the input file
		file, err := os.Open(inputFile)
		if err != nil {
			mergingErrs = append(mergingErrs, fmt.Errorf("failed to open %s: %w", inputFile, err))
			continue
		}

		fileInfo, err := file.Stat()
		if err != nil {
			mergingErrs = append(mergingErrs, fmt.Errorf("failed to stat file %s: %w", inputFile, err))
			file.Close()
			continue
		}

		if fileInfo.Size() == 0 {
			// Skip empty files
			log.Debug().Msgf("Skipped empty file: %s", inputFile)
			file.Close()
			continue
		}

		// Create the PCAP reader for the input file
		reader, err := pcapgo.NewReader(file)
		if err != nil {
			mergingErrs = append(mergingErrs, fmt.Errorf("failed to create pcapng reader for %v: %w", file.Name(), err))
			file.Close()
			continue
		}

		for {
			// Read packet data
			data, ci, err := reader.ReadPacketData()
			if err != nil {
				if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
					break
				}
				mergingErrs = append(mergingErrs, fmt.Errorf("error reading packet from file %s: %w", file.Name(), err))
				break
			}

			// Write the packet to the output file
			err = writer.WritePacket(ci, data)
			if err != nil {
				log.Error().Err(err).Msgf("Error writing packet to output file")
				mergingErrs = append(mergingErrs, fmt.Errorf("error writing packet to output file: %w", err))
				break
			}
		}

		file.Close()
	}

	log.Debug().Err(errors.Join(mergingErrs...))

	return nil
}

func copyPcapFiles(clientset *kubernetes.Clientset, config *rest.Config, destDir string, cutoffTime *time.Time) error {
	// List all namespaces
	namespaceList, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	var targetNamespaces []string
	for _, ns := range namespaceList.Items {
		targetNamespaces = append(targetNamespaces, ns.Name)
	}

	// List all worker pods
	workerPods, err := listWorkerPods(context.Background(), clientset, targetNamespaces)
	if err != nil {
		if len(workerPods) == 0 {
			return err
		}
		log.Debug().Err(err).Msg("error while listing worker pods")
	}

	var wg sync.WaitGroup

	// Launch a goroutine for each pod
	for _, pod := range workerPods {
		wg.Add(1)

		go func(pod *PodFileInfo) {
			defer wg.Done()

			// List files for the current pod
			err := listFilesInPodDir(context.Background(), clientset, config, pod, cutoffTime)
			if err != nil {
				log.Debug().Err(err).Msgf("error listing files in pod %s", pod.Pod.Name)
				return
			}

			// Copy files from the pod
			for _, file := range pod.Files {
				destFile := filepath.Join(destDir, file)

				// Add a timeout context for file copy
				ctx, cancel := context.WithTimeout(context.Background(), maxTimePerFile)
				err := copyFileFromPod(ctx, clientset, config, pod, file, destFile)
				cancel()
				if err != nil {
					log.Debug().Err(err).Msgf("error copying file %s from pod %s in namespace %s", file, pod.Pod.Name, pod.Pod.Namespace)
					continue
				}

				log.Info().Msgf("Copied file %s from pod %s to %s", file, pod.Pod.Name, destFile)
				pod.CopiedFiles = append(pod.CopiedFiles, destFile)
			}
		}(pod)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	var copiedFiles []string
	for _, pod := range workerPods {
		copiedFiles = append(copiedFiles, pod.CopiedFiles...)
	}

	if len(copiedFiles) == 0 {
		log.Info().Msg("No pcaps available to copy on the workers")
		return nil
	}

	// Generate a temporary filename for the merged file
	tempMergedFile := copiedFiles[0] + "_temp"

	// Merge PCAP files
	err = mergePCAPs(tempMergedFile, copiedFiles)
	if err != nil {
		os.Remove(tempMergedFile)
		return fmt.Errorf("error merging files: %w", err)
	}

	// Remove the original files after merging
	for _, file := range copiedFiles {
		if err := os.Remove(file); err != nil {
			log.Debug().Err(err).Msgf("error removing file %s", file)
		}
	}

	// Rename the temp file to the final name
	finalMergedFile := strings.TrimSuffix(tempMergedFile, "_temp")
	err = os.Rename(tempMergedFile, finalMergedFile)
	if err != nil {
		return err
	}

	log.Info().Msgf("Merged file created: %s", finalMergedFile)
	return nil
}
