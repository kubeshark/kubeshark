package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/wait"
	"k8s.io/client-go/util/retry"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/util/wait"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	kubesharkNamespace = "kubeshark"
	kubesharkLabel     = "app=kubeshark-monitored"
	podName            = "example-pod"
	image              = "nginx"
)

func main() {
	// Load Kubernetes configuration
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if err != nil {
		panic(err.Error())
	}

	// Create a Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Deploy Kubeshark (assuming you have the Kubeshark YAMLs in the 'kubeshark' directory)
	if err := deployKubeshark(clientset); err != nil {
		panic(fmt.Errorf("failed to deploy Kubeshark: %v", err))
	}

	// Label pods to be monitored (assuming you want to monitor pods with label 'app=kubeshark-monitored')
	if err := labelMonitoredPods(clientset); err != nil {
		panic(fmt.Errorf("failed to label monitored pods: %v", err))
	}

	// Create a Pod with lifecycle hooks (modify image and hooks as needed)
	if err := createPodWithLifecycleHooks(clientset, podName, image); err != nil {
		panic(fmt.Errorf("failed to create Pod with lifecycle hooks: %v", err))
	}

	// Subscribe to pod events
	if err := subscribeToPodEvents(clientset); err != nil {
		panic(fmt.Errorf("failed to subscribe to pod events: %v", err))
	}
}

func deployKubeshark(clientset kubernetes.Interface) error {
	// Implement Kubeshark deployment logic here (e.g., using clientset.AppsV1().Deployments())
	fmt.Println("Deploying Kubeshark...")
	// Add deployment logic here and handle errors
	return nil
}

func labelMonitoredPods(clientset kubernetes.Interface) error {
	// Implement pod labeling logic here (e.g., using clientset.CoreV1().Pods())
	fmt.Println("Labeling monitored pods...")
	// Add pod labeling logic here and handle errors
	return nil
}

func createPodWithLifecycleHooks(clientset kubernetes.Interface, podName, image string) error {
	// Implement Pod creation with lifecycle hooks (e.g., using clientset.CoreV1().Pods())
	fmt.Println("Creating Pod with lifecycle hooks...")
	// Add Pod creation with lifecycle hooks logic here and handle errors
	return nil
}

func handlePodEvent(event *v1.Event) {
	// Implement your logic to handle pod events here
	fmt.Printf("Received pod event: %s\n", event.Message)
}

func subscribeToPodEvents(clientset kubernetes.Interface) error {
	eventClient := clientset.CoreV1().Events(v1.NamespaceAll)
	ctx := context.TODO()

	// Watch for pod events using a list-watch
	listWatch := cache.NewListWatchFromClient(eventClient.RESTClient(), "events", v1.NamespaceAll, fields.Everything())

	_, controller := cache.NewInformer(
		listWatch,
		&v1.Event{},
		0, // Duration to resync. Use 0 for no resync.
		cache.ResourceEventHandlerFuncs{
			AddFunc:    handlePodEvent,
			UpdateFunc: handlePodEvent,
			DeleteFunc: handlePodEvent,
		},
	)

	stop := make(chan struct{})
	go func() {
		defer close(stop)
		controller.Run(stop)
	}()

	// Wait for the controller to stop or for an error to occur
	select {
	case <-stop:
		fmt.Println("Pod event subscription stopped.")
	case err := <-controller.Error():
		return fmt.Errorf("error in pod event controller: %v", err)
	}

	return nil
}
