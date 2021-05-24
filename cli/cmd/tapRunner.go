package cmd

import (
	"context"
	"fmt"
	"github.com/up9inc/mizu/cli/kubernetes"
	"github.com/up9inc/mizu/cli/mizu"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"
)

func RunMizuTap(podRegexQuery *regexp.Regexp, tappingOptions *MizuTapOptions) {
	kubernetesProvider := kubernetes.NewProvider(tappingOptions.KubeConfigPath, tappingOptions.Namespace)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	nodeToTappedPodIPMap, err := getNodeHostToTappedPodIpsMap(ctx, kubernetesProvider, podRegexQuery)
	if err != nil {
		cleanUpMizuResources(kubernetesProvider)
		return
	}
	err = createMizuResources(ctx, kubernetesProvider, nodeToTappedPodIPMap, tappingOptions)
	if err != nil {
		cleanUpMizuResources(kubernetesProvider)
		return
	}
	go portForwardApiPod(ctx, kubernetesProvider, cancel, tappingOptions) //TODO convert this to job for built in pod ttl or have the running app handle this
	waitForFinish(ctx, cancel)                                            //block until exit signal or error

	// TODO handle incoming traffic from tapper using a channel

	//cleanup
	fmt.Printf("\nRemoving mizu resources\n")
	cleanUpMizuResources(kubernetesProvider)
}

func createMizuResources(ctx context.Context, kubernetesProvider *kubernetes.Provider, nodeToTappedPodIPMap map[string][]string, tappingOptions *MizuTapOptions) error {
	mizuServiceAccountExists := createRBACIfNecessary(ctx, kubernetesProvider)
	_, err := kubernetesProvider.CreateMizuAggregatorPod(ctx, mizu.ResourcesNamespace, mizu.AggregatorPodName, tappingOptions.MizuImage, mizuServiceAccountExists)
	if err != nil {
		fmt.Printf("Error creating mizu collector pod: %v\n", err)
		return err
	}
	aggregatorService, err := kubernetesProvider.CreateService(ctx, mizu.ResourcesNamespace, mizu.AggregatorPodName, mizu.AggregatorPodName)
	if err != nil {
		fmt.Printf("Error creating mizu collector service: %v\n", err)
		return err
	}
	err = kubernetesProvider.CreateMizuTapperDaemonSet(ctx, mizu.ResourcesNamespace, mizu.TapperDaemonSetName, tappingOptions.MizuImage, mizu.TapperPodName, fmt.Sprintf("%s.%s.svc.cluster.local", aggregatorService.Name, aggregatorService.Namespace), nodeToTappedPodIPMap, mizuServiceAccountExists)
	if err != nil {
		fmt.Printf("Error creating mizu tapper daemonset: %v\n", err)
		return err
	}
	return nil
}

func cleanUpMizuResources(kubernetesProvider *kubernetes.Provider) {
	removalCtx, _ := context.WithTimeout(context.Background(), 5 * time.Second)
	if err := kubernetesProvider.RemovePod(removalCtx, mizu.ResourcesNamespace, mizu.AggregatorPodName); err != nil {
		fmt.Printf("Error removing Pod %s in namespace %s: %s (%v,%+v)\n", mizu.AggregatorPodName, mizu.ResourcesNamespace, err, err, err);
	}
	if err := kubernetesProvider.RemoveService(removalCtx, mizu.ResourcesNamespace, mizu.AggregatorPodName); err != nil {
		fmt.Printf("Error removing Service %s in namespace %s: %s (%v,%+v)\n", mizu.AggregatorPodName, mizu.ResourcesNamespace, err, err, err);
	}
	if err := kubernetesProvider.RemoveDaemonSet(removalCtx, mizu.ResourcesNamespace, mizu.TapperDaemonSetName); err != nil {
		fmt.Printf("Error removing DaemonSet %s in namespace %s: %s (%v,%+v)\n", mizu.TapperDaemonSetName, mizu.ResourcesNamespace, err, err, err);
	}
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

func portForwardApiPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc, tappingOptions *MizuTapOptions) {
	podExactRegex := regexp.MustCompile(fmt.Sprintf("^%s$", mizu.AggregatorPodName))
	added, modified, removed, errorChan := kubernetes.FilteredWatch(ctx, kubernetesProvider.GetPodWatcher(ctx, mizu.ResourcesNamespace), podExactRegex)
	isPodReady := false
	var portForward *kubernetes.PortForward
	for {
		select {
		case <- added:
			continue
		case <- removed:
			fmt.Printf("%s removed\n", mizu.AggregatorPodName)
			cancel()
			return
		case modifiedPod := <- modified:
			if modifiedPod.Status.Phase == "Running" && !isPodReady {
				isPodReady = true
				var err error
				portForward, err = kubernetes.NewPortForward(kubernetesProvider, mizu.ResourcesNamespace, mizu.AggregatorPodName, tappingOptions.GuiPort, tappingOptions.MizuPodPort, cancel)
				fmt.Printf("Web interface is now available at http://localhost:%d\n", tappingOptions.GuiPort)
				if err != nil {
					fmt.Printf("error forwarding port to pod %s\n", err)
					cancel()
				}
			}

		case <- time.After(25 * time.Second):
			if !isPodReady {
				fmt.Printf("error: %s pod was not ready in time", mizu.AggregatorPodName)
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
	mizuRBACExists, err := kubernetesProvider.DoesMizuRBACExist(ctx, mizu.ResourcesNamespace)
	if err != nil {
		fmt.Printf("warning: could not ensure mizu rbac resources exist %v\n", err)
		return false
	}
	if !mizuRBACExists {
		var versionString = mizu.Version
		if mizu.GitCommitHash != "" {
			versionString += "-" + mizu.GitCommitHash
		}
		err := kubernetesProvider.CreateMizuRBAC(ctx, mizu.ResourcesNamespace, versionString)
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


