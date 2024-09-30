package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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

// listWorkerPods fetches all the worker pods using the Kubernetes client
func listWorkerPods(ctx context.Context, clientset *clientk8s.Clientset, namespace string) (*corev1.PodList, error) {
	labelSelector := label

	// List all pods matching the label
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list worker pods: %w", err)
	}

	return pods, nil
}

// listFilesInPodDir lists all files in the specified directory inside the pod
func listFilesInPodDir(ctx context.Context, clientset *clientk8s.Clientset, config *rest.Config, podName, namespace, srcDir string) ([]string, error) {
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
		return nil, fmt.Errorf("failed to initialize executor: %w", err)
	}

	// Buffer to capture stdout (file listing)
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	// Stream the result of the ls command
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdoutBuf,
		Stderr: &stderrBuf, // Capture stderr for better debugging
	})
	if err != nil {
		return nil, fmt.Errorf("error listing files in pod: %w. Stderr: %s", err, stderrBuf.String())
	}

	// Split the output (file names) into a list
	files := strings.Split(strings.TrimSpace(stdoutBuf.String()), "\n")
	return files, nil
}

// copyFileFromPod copies a single file from a pod to a local destination
func copyFileFromPod(ctx context.Context, clientset *clientk8s.Clientset, config *rest.Config, podName, namespace, srcFile, destFile string) error {
	cmd := []string{"cat", srcFile}
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
		return fmt.Errorf("failed to initialize executor: %w", err)
	}

	// Create a local file to write the content to
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
		Stderr: &stderrBuf, // Capture stderr for better debugging
	})
	if err != nil {
		return fmt.Errorf("error copying file from pod: %w. Stderr: %s", err, stderrBuf.String())
	}

	return nil
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

func setPcapConfigInKubernetes(clientset *clientk8s.Clientset, podName, namespace, enabledPcap, timeInterval, maxTime, maxSize string) error {
	// Load the existing ConfigMap
	configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), "kubeshark-config-map", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to load ConfigMap: %w", err)
	}

	// Update the values with user-provided input
	if len(configMap.Data["PCAP_TIME_INTERVAL"]) > 0 {
		configMap.Data["PCAP_TIME_INTERVAL"] = timeInterval

	}
	if len(configMap.Data["PCAP_MAX_SIZE"]) > 0 {
		configMap.Data["PCAP_MAX_SIZE"] = maxSize

	}
	if len(configMap.Data["PCAP_MAX_TIME"]) > 0 {
		configMap.Data["PCAP_MAX_TIME"] = maxTime

	}
	if len(configMap.Data["PCAP_DUMP_ENABLE"]) > 0 {
		configMap.Data["PCAP_DUMP_ENABLE"] = enabledPcap
	}

	// Apply the updated ConfigMap back to the cluster
	_, err = clientset.CoreV1().ConfigMaps(namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ConfigMap: %w", err)
	}

	return nil
}

// WorkerSrcResponse represents the response structure from the worker's /pcaps/worker-src endpoint.
type WorkerSrcResponse struct {
	WorkerSrcDir string `json:"workerSrcDir"`
}

// getWorkerSource fetches the worker source directory from the worker pod via the /pcaps/worker-src endpoint.
func getWorkerSource(clientset *kubernetes.Clientset, podName, namespace string) (string, error) {
	// Get the worker pod IP or service address (you can also use the cluster DNS name)
	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		log.Error().Err(err).Msgf("Failed to get pod %s", podName)
		return "", err
	}

	// Construct the URL to access the worker's /pcaps/worker-src endpoint
	workerURL := fmt.Sprintf("http://%s:30001/pcaps/worker-src", pod.Status.PodIP)

	// Make an HTTP request to the worker pod's endpoint
	resp, err := http.Get(workerURL)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to request worker src dir from %s", workerURL)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get worker src dir, status code: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse the JSON response
	var workerSrcResp WorkerSrcResponse
	err = json.Unmarshal(body, &workerSrcResp)
	if err != nil {
		return "", fmt.Errorf("failed to parse worker src dir response: %v", err)
	}

	return workerSrcResp.WorkerSrcDir, nil
}
