package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

const label = "app.kubeshark.co/app=worker"

// listWorkerPods fetches all the worker pods using the Kubernetes client
func listWorkerPods(ctx context.Context, clientset *kubernetes.Clientset, namespace string) (*corev1.PodList, error) {
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
func listFilesInPodDir(ctx context.Context, clientset *kubernetes.Clientset, config *rest.Config, podName, namespace, srcDir string) ([]string, error) {
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
func copyFileFromPod(ctx context.Context, clientset *kubernetes.Clientset, config *rest.Config, podName, namespace, srcFile, destFile string) error {
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

	fmt.Printf("File from pod %s copied to local destination: %s\n", podName, destFile)
	return nil
}

// updatePodEnvVars updates the configuration file inside the worker pod
func updatePodEnvVars(ctx context.Context, clientset *kubernetes.Clientset, config *rest.Config, podName, namespace string, stop bool, timeInterval, maxTime, maxSize string) error {
	var envVars []string
	if stop {
		envVars = append(envVars, "PCAP_DUMP_ENABLE=false")
	} else {
		envVars = append(envVars, "PCAP_DUMP_ENABLE=true")

		if timeInterval != "" {
			envVars = append(envVars, fmt.Sprintf("TIME_INTERVAL=%s", timeInterval))
		}
		if maxTime != "" {
			envVars = append(envVars, fmt.Sprintf("MAX_TIME=%s", maxTime))
		}
		if maxSize != "" {
			envVars = append(envVars, fmt.Sprintf("MAX_SIZE=%s", maxSize))
		}
	}

	// Create a command that sets the environment variables directly in the pod
	for _, envVar := range envVars {
		cmd := []string{"sh", "-c", fmt.Sprintf("export %s", envVar)}
		req := clientset.CoreV1().RESTClient().Post().
			Resource("pods").
			Name(podName).
			Namespace(namespace).
			SubResource("exec").
			Param("container", "sniffer"). // Assuming container is called 'sniffer'
			Param("stdout", "true").
			Param("stderr", "true").
			Param("command", cmd[0]).
			Param("command", cmd[1]).
			Param("command", cmd[2])

		exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
		if err != nil {
			return fmt.Errorf("failed to initialize executor for pod %s: %w", podName, err)
		}

		var stdoutBuf, stderrBuf bytes.Buffer
		err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
			Stdout: &stdoutBuf,
			Stderr: &stderrBuf,
		})
		if err != nil {
			return fmt.Errorf("failed to update env vars in pod %s: %w. Stderr: %s", podName, err, stderrBuf.String())
		}
	}

	fmt.Printf("Updated env vars for pod %s\n", podName)
	return nil
}

// readConfigFileFromPod reads the configuration file from the pod
func readConfigFileFromPod(ctx context.Context, clientset *kubernetes.Clientset, config *rest.Config, podName, namespace, configFilePath string) (map[string]string, error) {
	cmd := []string{"cat", configFilePath}
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
		return nil, fmt.Errorf("failed to initialize executor for pod %s: %w", podName, err)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdoutBuf,
		Stderr: &stderrBuf,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read config file from pod %s: %w. Stderr: %s", podName, err, stderrBuf.String())
	}

	// Parse the config content into a map of key-value pairs
	configMap := parseConfigContent(stdoutBuf.String())
	return configMap, nil
}

// writeConfigFileToPod writes the updated configuration map to the file in the pod
func writeConfigFileToPod(ctx context.Context, clientset *kubernetes.Clientset, config *rest.Config, podName, namespace, configFilePath string, configMap map[string]string) error {
	// Convert the config map back to a string format for writing
	configContent := formatConfigMapToString(configMap)

	// Escape any single quotes in the config content to avoid issues in the shell command
	escapedConfigContent := strings.ReplaceAll(configContent, "'", "'\\''")

	// Prepare the command to write the configuration to the file
	cmd := []string{"sh", "-c", fmt.Sprintf("echo '%s' > %s", escapedConfigContent, configFilePath)}
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		Param("container", "sniffer").
		Param("stdout", "true").
		Param("stderr", "true").
		Param("command", cmd[0]).
		Param("command", cmd[1]).
		Param("command", cmd[2])

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("failed to initialize executor for pod %s: %w", podName, err)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdoutBuf,
		Stderr: &stderrBuf,
	})
	if err != nil {
		return fmt.Errorf("failed to write config file to pod %s: %w. Stderr: %s", podName, err, stderrBuf.String())
	}

	return nil
}

// parseConfigContent parses the content of the config file into a map of key-value pairs
func parseConfigContent(content string) map[string]string {
	configMap := make(map[string]string)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" || !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		configMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return configMap
}

// formatConfigMapToString converts the config map back to string format
func formatConfigMapToString(configMap map[string]string) string {
	var sb strings.Builder
	for key, value := range configMap {
		sb.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	}
	return sb.String()
}
