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

func Run(podRegexQuery *regexp.Regexp) {
	kubernetesProvider := kubernetes.NewProvider(config.Configuration.KubeConfigPath, config.Configuration.Namespace)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	nodeToTappedPodIPMap, err := getNodeHostToTappedPodIpsMap(ctx, kubernetesProvider, podRegexQuery)
	if err != nil {
		cleanUpMizuResources(kubernetesProvider)
		return
	}
	err = createMizuResources(ctx, kubernetesProvider, nodeToTappedPodIPMap)
	if err != nil {
		cleanUpMizuResources(kubernetesProvider)
		return
	}
	go portForwardApiPod(ctx, kubernetesProvider, cancel) //TODO convert this to job for built in pod ttl or have the running app handle this
	waitForFinish(ctx, cancel)                                                                                                                //block until exit signal or error

	// TODO handle incoming traffic from tapper using a channel

	//cleanup
	fmt.Printf("\nRemoving mizu resources\n")
	cleanUpMizuResources(kubernetesProvider)
}

func createMizuResources(ctx context.Context, kubernetesProvider *kubernetes.Provider, nodeToTappedPodIPMap map[string][]string) error {
	mizuServiceAccountExists := createRBACIfNecessary(ctx, kubernetesProvider)
	_, err := kubernetesProvider.CreateMizuAggregatorPod(ctx, MizuResourcesNamespace, aggregatorPodName, config.Configuration.MizuImage, mizuServiceAccountExists)
	if err != nil {
		fmt.Printf("Error creating mizu collector pod: %v\n", err)
		return err
	}
	aggregatorService, err := kubernetesProvider.CreateService(ctx, MizuResourcesNamespace, aggregatorPodName, aggregatorPodName)
	if err != nil {
		fmt.Printf("Error creating mizu collector service: %v\n", err)
		return err
	}
	err = kubernetesProvider.CreateMizuTapperDaemonSet(ctx, MizuResourcesNamespace, TapperDaemonSetName, config.Configuration.MizuImage, tapperPodName, fmt.Sprintf("%s.%s.svc.cluster.local", aggregatorService.Name, aggregatorService.Namespace), nodeToTappedPodIPMap, mizuServiceAccountExists)
	if err != nil {
		fmt.Printf("Error creating mizu tapper daemonset: %v\n", err)
		return err
	}
	return nil
}

func cleanUpMizuResources(kubernetesProvider *kubernetes.Provider) {
	removalCtx, _ := context.WithTimeout(context.Background(), 5 * time.Second)
	kubernetesProvider.RemovePod(removalCtx, MizuResourcesNamespace, aggregatorPodName)
	kubernetesProvider.RemoveService(removalCtx, MizuResourcesNamespace, aggregatorPodName)
	kubernetesProvider.RemoveDaemonSet(removalCtx, MizuResourcesNamespace, TapperDaemonSetName)
}

// will be relevant in the future
//func watchPodsForTapping(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc, podRegex *regexp.Regexp) {
//	added, modified, removed, errorChan := kubernetes.FilteredWatch(ctx, kubernetesProvider.GetPodWatcher(ctx, kubernetesProvider.Namespace), podRegex)
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

func portForwardApiPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc) {
	podExactRegex := regexp.MustCompile(fmt.Sprintf("^%s$", aggregatorPodName))
	added, modified, removed, errorChan := kubernetes.FilteredWatch(ctx, kubernetesProvider.GetPodWatcher(ctx, MizuResourcesNamespace), podExactRegex)
	isPodReady := false
	var portForward *kubernetes.PortForward
	for {
		select {
		case <- added:
			continue
		case <- removed:
			fmt.Printf("%s removed\n", aggregatorPodName)
			cancel()
			return
		case modifiedPod := <- modified:
			if modifiedPod.Status.Phase == "Running" && !isPodReady {
				isPodReady = true
				var err error
				portForward, err = kubernetes.NewPortForward(kubernetesProvider, MizuResourcesNamespace, aggregatorPodName, config.Configuration.GuiPort, config.Configuration.MizuPodPort, cancel)
				fmt.Printf("Web interface is now available at http://localhost:%d\n", config.Configuration.GuiPort)
				if err != nil {
					fmt.Printf("error forwarding port to pod %s\n", err)
					cancel()
				}
			}

		case <- time.After(25 * time.Second):
			if !isPodReady {
				fmt.Printf("error: %s pod was not ready in time", aggregatorPodName)
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

func getNodeHostToTappedPodIpsMap(ctx context.Context, kubernetesProvider *kubernetes.Provider, regex *regexp.Regexp) (map[string][]string, error) {
	matchingPods, err := kubernetesProvider.GetAllPodsMatchingRegex(ctx, regex)
	if err != nil {
		return nil, err
	}
	nodeToTappedPodIPMap := make(map[string][]string, 0)
	for _, pod := range matchingPods {
		existingList := nodeToTappedPodIPMap[pod.Spec.NodeName]
		if existingList == nil {
			nodeToTappedPodIPMap[pod.Spec.NodeName] = []string {pod.Status.PodIP}
		} else {
			nodeToTappedPodIPMap[pod.Spec.NodeName] = append(nodeToTappedPodIPMap[pod.Spec.NodeName], pod.Status.PodIP)
		}
	}
	return nodeToTappedPodIPMap, nil
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


