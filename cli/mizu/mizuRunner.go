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

func Run(tappedPodName string) {
	kubernetesProvider := kubernetes.NewProvider(config.Configuration.KubeConfigPath, config.Configuration.Namespace)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	podName := "mizu-collector"

	mizuServiceAccountExists := createRBACIfNecessary(ctx, kubernetesProvider)
	go createPodAndPortForward(ctx, kubernetesProvider, cancel, podName, MizuResourcesNamespace, tappedPodName, mizuServiceAccountExists) //TODO convert this to job for built in pod ttl or have the running app handle this
	waitForFinish(ctx, cancel) //block until exit signal or error

	// TODO handle incoming traffic from tapper using a channel

	//cleanup
	fmt.Printf("\nremoving pod %s\n", podName)
	removalCtx, _ := context.WithTimeout(context.Background(), 2 * time.Second)
	kubernetesProvider.RemovePod(removalCtx, MizuResourcesNamespace, podName)
}

func watchPodsForTapping(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc, podRegex *regexp.Regexp) {
	added, modified, removed, errorChan := kubernetes.FilteredWatch(ctx, kubernetesProvider.GetPodWatcher(ctx, kubernetesProvider.Namespace), podRegex)
	for {
		select {
		case newTarget := <- added:
			fmt.Printf("+%s\n", newTarget.Name)

		case removedTarget := <- removed:
			fmt.Printf("-%s\n", removedTarget.Name)

		case <- modified:
			continue

		case <- errorChan:
			cancel()

		case <- ctx.Done():
			return
		}
	}
}

func createPodAndPortForward(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc, podName string, namespace string, tappedPodName string, linkServiceAccount bool) {
	pod, err := kubernetesProvider.CreateMizuPod(ctx, MizuResourcesNamespace, podName, config.Configuration.MizuImage, kubernetesProvider.Namespace, tappedPodName, linkServiceAccount)
	if err != nil {
		fmt.Printf("error creating pod %s", err)
		cancel()
		return
	}
	podExactRegex := regexp.MustCompile(fmt.Sprintf("^%s$", pod.Name))
	added, modified, removed, errorChan := kubernetes.FilteredWatch(ctx, kubernetesProvider.GetPodWatcher(ctx, namespace), podExactRegex)
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
				portForward, err = kubernetes.NewPortForward(kubernetesProvider, namespace, podName, config.Configuration.GuiPort, config.Configuration.MizuPodPort, cancel)
				fmt.Printf("Web interface is now available at http://localhost:%d\n", config.Configuration.GuiPort)
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

func createRBACIfNecessary(ctx context.Context, kubernetesProvider *kubernetes.Provider) bool {
	mizuRBACExists, err := kubernetesProvider.DoesMizuRBACExist(ctx, MizuResourcesNamespace)
	if err != nil {
		fmt.Printf("warning: could not ensure mizu rbac resources exist %v\n", err)
		return false
	}
	if !mizuRBACExists {
		var versionString = Version
		if GitCommitHash != "" {
			versionString += "-" + GitCommitHash
		}
		err := kubernetesProvider.CreateMizuRBAC(ctx, MizuResourcesNamespace, versionString)
		if err != nil {
			fmt.Printf("warning: could not create mizu rbac resources %v\n", err)
			return false
		}
	}
	return true
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


