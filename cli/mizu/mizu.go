package mizu

import (
	"context"
	"fmt"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/kubernetes"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"
)

var (
	isPortForwarded = false
)

func Run(podRegex *regexp.Regexp) {
	kubernetesProvider := kubernetes.NewProvider(config.Configuration.KubeConfigPath, config.Configuration.Namespace)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	podName := "mizu-collector"

	go createPodAndPortForward(ctx, kubernetesProvider, cancel, podName) //TODO convert this to job for built in pod ttl or have the running app handle this
	go watchPodsForTapping(ctx, kubernetesProvider, cancel, podRegex)
	waitForFinish(ctx, cancel)

	// TODO handle incoming traffic from tapper using a channel

	fmt.Printf("\nremoving pod %s\n", podName)
	removalCtx, _ := context.WithTimeout(context.Background(), 2 * time.Second)
	kubernetesProvider.RemovePod(removalCtx, podName)
}

func watchPodsForTapping(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc, podRegex *regexp.Regexp) {
	added, _, removed, errorChan := kubernetes.FilteredWatch(ctx, kubernetesProvider.GetPodWatcher(ctx), podRegex)
	for {
		select {
		case newTarget := <- added:
			fmt.Printf("+%s\n", newTarget.Name)

		case removedTarget := <- removed:
			fmt.Printf("-%s\n", removedTarget.Name)

		case <- errorChan:
			cancel()

		case <- ctx.Done():
			return
		}
	}
}

func createPodAndPortForward(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc, podName string) {
	podImage := "kennethreitz/httpbin:latest"

	pod, err := kubernetesProvider.CreatePod(ctx, podName, podImage)
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
				portForward, err = kubernetes.NewPortForward(kubernetesProvider, kubernetesProvider.Namespace, podName, config.Configuration.DashboardPort, 80, cancel)
				if !config.Configuration.NoDashboard {
					fmt.Printf("Dashboard is now available at http://localhost:%d\n", config.Configuration.DashboardPort)
				}
				if err != nil {
					fmt.Printf("error forwarding port to pod %s\n", err)
					cancel()
				}
			}

		case <- time.After(10 * time.Second):
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


