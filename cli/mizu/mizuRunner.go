package mizu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/kubernetes"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"
)

func Run() {
	kubernetesProvider := kubernetes.NewProvider(config.Configuration.KubeConfigPath, config.Configuration.Namespace)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	podName := "mizu-collector"

	go createPodAndPortForward(ctx, kubernetesProvider, cancel, podName) //TODO convert this to job for built in pod ttl or have the running app handle this
	waitForFinish(ctx, cancel) //block until exit signal or error

	// TODO handle incoming traffic from tapper using a channel

	//cleanup
	fmt.Printf("\nremoving pod %s\n", podName)
	removalCtx, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
	defer cancel()
	kubernetesProvider.RemovePod(removalCtx, podName)
}

//func watchPodsForTapping(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc, podRegex *regexp.Regexp) {
//	added, modified, removed, errorChan := kubernetes.FilteredWatch(ctx, kubernetesProvider.GetPodWatcher(ctx), podRegex)
//	for {
//		select {
//		case newTarget := <- added:
//			fmt.Printf("+%s\n", newTarget.Name)
//
//		case removedTarget := <- removed:
//			fmt.Printf("-%s\n", removedTarget.Name)
//
//		case <- modified:
//			continue
//
//		case <- errorChan:
//			cancel()
//
//		case <- ctx.Done():
//			return
//		}
//	}
//}

type ChangeAddressesBody struct {
	Addresses []string `json:"addresses"`
}

func createPodAndPortForward(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc, podName string) {
	pod, err := kubernetesProvider.CreateMizuPod(ctx, podName, config.Configuration.MizuImage, config.Configuration.TappedPodName)
	if err != nil {
		fmt.Printf("error creating pod %s", err)
		cancel()
		return
	}
	podExactRegex := regexp.MustCompile(fmt.Sprintf("^%s$", pod.Name))
	added, modified, removed, errorChan := kubernetes.FilteredWatch(ctx, kubernetesProvider.GetPodWatcher(ctx), podExactRegex)
	isPodReady := false
	var portForward *kubernetes.PortForward
	for {
		select {
		case <- added:
			continue
		case <- removed:
			fmt.Printf("%s removed\n", podName)
			cancel()
			return
		case modifiedPod := <- modified:
			if modifiedPod.Status.Phase == "Running" && !isPodReady {
				isPodReady = true
				var err error
				portForward, err = kubernetes.NewPortForward(kubernetesProvider, kubernetesProvider.Namespace, podName, config.Configuration.DashboardPort, config.Configuration.MizuPodPort, cancel)

				sendAddressesToTapper(modifiedPod)

				if !config.Configuration.NoDashboard {
					fmt.Printf("Dashboard is now available at http://localhost:%d\n", config.Configuration.DashboardPort)
				}
				if err != nil {
					fmt.Printf("error forwarding port to pod %s\n", err)
					cancel()
				}
			}

		case <- time.After(25 * time.Second):
			if !isPodReady {
				fmt.Printf("error: %s pod was not ready in time", podName)
				cancel()
			}

		case <- errorChan:
			cancel()

		case <- ctx.Done():
			if portForward != nil {
				portForward.Stop()
			}
			return
		}
	}
}

func sendAddressesToTapper(pod *corev1.Pod) {
	addresses := getAddresses(pod)
	data := &ChangeAddressesBody{
		Addresses: addresses,
	}
	dataBytes, _ := json.Marshal(data)
	bytesReader := bytes.NewReader(dataBytes)
	_, _ = http.Post("http://localhost:8899/proxy/tapper", "application/json", bytesReader)
}

func getAddresses(tappedPod *corev1.Pod) []string {
	podIps := make([]string, len(tappedPod.Status.PodIPs))
	for ii, podIp := range tappedPod.Status.PodIPs {
		podIps[ii] = podIp.IP
	}
	return podIps
}

func waitForFinish(ctx context.Context, cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// block until ctx cancel is called or termination signal is received
	select {
	case <- ctx.Done():
		break
	case <- sigChan:
		cancel()
	}
}


