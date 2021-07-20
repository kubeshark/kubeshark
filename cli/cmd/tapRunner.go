package cmd

import (
	"context"
	"fmt"
	"github.com/romana/rlog"
	"github.com/up9inc/mizu/cli/debounce"
	"github.com/up9inc/mizu/cli/kubernetes"
	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/shared"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"
)

var mizuServiceAccountExists bool
var aggregatorService *core.Service

const (
	updateTappersDelay = 5 * time.Second
)

var currentlyTappedPods []core.Pod

func RunMizuTap(podRegexQuery *regexp.Regexp, tappingOptions *MizuTapOptions) {
	mizuApiFilteringOptions, err := getMizuApiFilteringOptions(tappingOptions)
	if err != nil {
		return
	}

	kubernetesProvider := kubernetes.NewProvider(tappingOptions.KubeConfigPath)

	defer cleanUpMizuResources(kubernetesProvider)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	targetNamespace := getNamespace(tappingOptions, kubernetesProvider)
	if matchingPods, err := kubernetesProvider.GetAllPodsMatchingRegex(ctx, podRegexQuery, targetNamespace); err != nil {
		return
	} else {
		currentlyTappedPods = matchingPods
	}

	var namespacesStr string
	if targetNamespace != mizu.K8sAllNamespaces {
		namespacesStr = fmt.Sprintf("namespace \"%s\"", targetNamespace)
	} else {
		namespacesStr = "all namespaces"
	}
	fmt.Printf("Tapping pods in %s\n", namespacesStr)

	if len(currentlyTappedPods) == 0 {
		var suggestionStr string
		if targetNamespace != mizu.K8sAllNamespaces {
			suggestionStr = "\nSelect a different namespace with -n or tap all namespaces with -A"
		}
		fmt.Printf("Did not find any pods matching the regex argument%s\n", suggestionStr)
	}

	nodeToTappedPodIPMap, err := getNodeHostToTappedPodIpsMap(currentlyTappedPods)
	if err != nil {
		return
	}

	if err := createMizuResources(ctx, kubernetesProvider, nodeToTappedPodIPMap, tappingOptions, mizuApiFilteringOptions); err != nil {
		return
	}

	go portForwardApiPod(ctx, kubernetesProvider, cancel, tappingOptions) // TODO convert this to job for built in pod ttl or have the running app handle this
	go watchPodsForTapping(ctx, kubernetesProvider, cancel, podRegexQuery, tappingOptions)
	go syncApiStatus(ctx, cancel, tappingOptions)

	//block until exit signal or error
	waitForFinish(ctx, cancel)
}

func createMizuResources(ctx context.Context, kubernetesProvider *kubernetes.Provider, nodeToTappedPodIPMap map[string][]string, tappingOptions *MizuTapOptions, mizuApiFilteringOptions *shared.TrafficFilteringOptions) error {
	if err := createMizuNamespace(ctx, kubernetesProvider); err != nil {
		return err
	}

	if err := createMizuAggregator(ctx, kubernetesProvider, tappingOptions, mizuApiFilteringOptions); err != nil {
		return err
	}

	if err := updateMizuTappers(ctx, kubernetesProvider, nodeToTappedPodIPMap, tappingOptions); err != nil {
		return err
	}

	return nil
}

func createMizuNamespace(ctx context.Context, kubernetesProvider *kubernetes.Provider) error {
	_, err := kubernetesProvider.CreateNamespace(ctx, mizu.ResourcesNamespace)
	if err != nil {
		fmt.Printf("Error creating Namespace %s: %v\n", mizu.ResourcesNamespace, err)
	}

	return err
}

func createMizuAggregator(ctx context.Context, kubernetesProvider *kubernetes.Provider, tappingOptions *MizuTapOptions, mizuApiFilteringOptions *shared.TrafficFilteringOptions) error {
	var err error

	mizuServiceAccountExists = createRBACIfNecessary(ctx, kubernetesProvider)
	var serviceAccountName string
	if mizuServiceAccountExists {
		serviceAccountName = mizu.ServiceAccountName
	} else {
		serviceAccountName = ""
	}
	_, err = kubernetesProvider.CreateMizuAggregatorPod(ctx, mizu.ResourcesNamespace, mizu.AggregatorPodName, tappingOptions.MizuImage, serviceAccountName, mizuApiFilteringOptions)
	if err != nil {
		fmt.Printf("Error creating mizu collector pod: %v\n", err)
		return err
	}

	aggregatorService, err = kubernetesProvider.CreateService(ctx, mizu.ResourcesNamespace, mizu.AggregatorPodName, mizu.AggregatorPodName)
	if err != nil {
		fmt.Printf("Error creating mizu collector service: %v\n", err)
		return err
	}

	return nil
}

func getMizuApiFilteringOptions(tappingOptions *MizuTapOptions) (*shared.TrafficFilteringOptions, error) {
	if tappingOptions.PlainTextFilterRegexes == nil || len(tappingOptions.PlainTextFilterRegexes) == 0 {
		return nil, nil
	}

	compiledRegexSlice := make([]*shared.SerializableRegexp, 0)
	for _, regexStr := range tappingOptions.PlainTextFilterRegexes {
		compiledRegex, err := shared.CompileRegexToSerializableRegexp(regexStr)
		if err != nil {
			fmt.Printf("Regex %s is invalid: %v", regexStr, err)
			return nil, err
		}
		compiledRegexSlice = append(compiledRegexSlice, compiledRegex)
	}

	return &shared.TrafficFilteringOptions{PlainTextMaskingRegexes: compiledRegexSlice}, nil
}

func updateMizuTappers(ctx context.Context, kubernetesProvider *kubernetes.Provider, nodeToTappedPodIPMap map[string][]string, tappingOptions *MizuTapOptions) error {
	if len(nodeToTappedPodIPMap) > 0 {
		var serviceAccountName string
		if mizuServiceAccountExists {
			serviceAccountName = mizu.ServiceAccountName
		} else {
			serviceAccountName = ""
		}

		if err := kubernetesProvider.ApplyMizuTapperDaemonSet(
			ctx,
			mizu.ResourcesNamespace,
			mizu.TapperDaemonSetName,
			tappingOptions.MizuImage,
			mizu.TapperPodName,
			fmt.Sprintf("%s.%s.svc.cluster.local", aggregatorService.Name, aggregatorService.Namespace),
			nodeToTappedPodIPMap,
			serviceAccountName,
			tappingOptions.TapOutgoing,
		); err != nil {
			fmt.Printf("Error creating mizu tapper daemonset: %v\n", err)
			return err
		}
	} else {
		if err := kubernetesProvider.RemoveDaemonSet(ctx, mizu.ResourcesNamespace, mizu.TapperDaemonSetName); err != nil {
			fmt.Printf("Error deleting mizu tapper daemonset: %v\n", err)
			return err
		}
	}

	return nil
}

func cleanUpMizuResources(kubernetesProvider *kubernetes.Provider) {
	fmt.Printf("\nRemoving mizu resources\n")

	removalCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Call cancel if a terminating signal was received
	go func() {
		waitForFinish(removalCtx, cancel)
	}()

	if err := kubernetesProvider.RemoveNamespace(removalCtx, mizu.ResourcesNamespace); err != nil {
		fmt.Printf("Error removing Namespace %s: %s (%v,%+v)\n", mizu.ResourcesNamespace, err, err, err)
		return
	}

	if err := kubernetesProvider.RemoveNonNamespacedResources(removalCtx, mizu.ClusterRoleName, mizu.ClusterRoleBindingName); err != nil {
		fmt.Printf("Error removing non-namespaced resources: %s (%v,%+v)\n", err, err, err)
		return
	}

	if err := kubernetesProvider.WaitUtilNamespaceDeleted(removalCtx, mizu.ResourcesNamespace); err != nil {
		switch {
		case removalCtx.Err() == context.Canceled:
			// Do nothing. User interrupted the wait.
		case err == wait.ErrWaitTimeout:
			fmt.Printf("Timeout while removing Namespace %s\n", mizu.ResourcesNamespace)
		default:
			fmt.Printf("Error while waiting for Namespace %s to be deleted: %s (%v,%+v)\n", mizu.ResourcesNamespace, err, err, err)
		}
	}
}

func watchPodsForTapping(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc, podRegex *regexp.Regexp, tappingOptions *MizuTapOptions) {
	targetNamespace := getNamespace(tappingOptions, kubernetesProvider)

	added, modified, removed, errorChan := kubernetes.FilteredWatch(ctx, kubernetesProvider.GetPodWatcher(ctx, targetNamespace), podRegex)

	restartTappers := func() {
		if matchingPods, err := kubernetesProvider.GetAllPodsMatchingRegex(ctx, podRegex, targetNamespace); err != nil {
			fmt.Printf("Error getting pods by regex: %s (%v,%+v)\n", err, err, err)
			cancel()
		} else {
			currentlyTappedPods = matchingPods
		}

		nodeToTappedPodIPMap, err := getNodeHostToTappedPodIpsMap(currentlyTappedPods)
		if err != nil {
			fmt.Printf("Error building node to ips map: %s (%v,%+v)\n", err, err, err)
			cancel()
		}

		if err := updateMizuTappers(ctx, kubernetesProvider, nodeToTappedPodIPMap, tappingOptions); err != nil {
			fmt.Printf("Error updating daemonset: %s (%v,%+v)\n", err, err, err)
			cancel()
		}
	}
	restartTappersDebouncer := debounce.NewDebouncer(updateTappersDelay, restartTappers)

	for {
		select {
		case newTarget := <-added:
			fmt.Printf(mizu.Green, fmt.Sprintf("+%s\n", newTarget.Name))

		case removedTarget := <-removed:
			fmt.Printf(mizu.Red, fmt.Sprintf("-%s\n", removedTarget.Name))
			restartTappersDebouncer.SetOn()

		case modifiedTarget := <-modified:
			// Act only if the modified pod has already obtained an IP address.
			// After filtering for IPs, on a normal pod restart this includes the following events:
			// - Pod deletion
			// - Pod reaches start state
			// - Pod reaches ready state
			// Ready/unready transitions might also trigger this event.
			if modifiedTarget.Status.PodIP != "" {
				restartTappersDebouncer.SetOn()
			}

		case <-errorChan:
			// TODO: Does this also perform cleanup?
			cancel()

		case <-ctx.Done():
			return
		}
	}
}

func portForwardApiPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc, tappingOptions *MizuTapOptions) {
	podExactRegex := regexp.MustCompile(fmt.Sprintf("^%s$", mizu.AggregatorPodName))
	added, modified, removed, errorChan := kubernetes.FilteredWatch(ctx, kubernetesProvider.GetPodWatcher(ctx, mizu.ResourcesNamespace), podExactRegex)
	isPodReady := false
	for {
		select {
		case <-ctx.Done():
			return

		case <-added:
			continue
		case <-removed:
			fmt.Printf("%s removed\n", mizu.AggregatorPodName)
			cancel()
			return
		case modifiedPod := <-modified:
			if modifiedPod.Status.Phase == "Running" && !isPodReady {
				isPodReady = true
				go func() {
					err := kubernetes.StartProxy(kubernetesProvider, tappingOptions.GuiPort, mizu.ResourcesNamespace, mizu.AggregatorPodName)
					if err != nil {
						fmt.Printf("Error occured while running k8s proxy %v\n", err)
						cancel()
					}
				}()
				mizuProxiedUrl := kubernetes.GetMizuCollectorProxiedHostAndPath(tappingOptions.GuiPort, mizu.ResourcesNamespace, mizu.AggregatorPodName)
				fmt.Printf("Mizu is available at  http://%s\n", mizuProxiedUrl)

				time.Sleep(time.Second * 5) // Waiting to be sure the proxy is ready
				if tappingOptions.Analyze {
					url_path := fmt.Sprintf("http://%s/api/uploadEntries?dest=%s", mizuProxiedUrl, url.QueryEscape(tappingOptions.AnalyzeDestination))
					u, err := url.ParseRequestURI(url_path)
					if err != nil {
						log.Fatal(fmt.Sprintf("Failed parsing the URL %v\n", err))
					}
					rlog.Debugf("Sending get request to %v\n", u.String())
					if response, err := http.Get(u.String()); err != nil && response.StatusCode != 200 {
						fmt.Printf("error sending upload entries req %v\n", err)
					} else {
						fmt.Printf(mizu.Purple, "Traffic is uploading to UP9 for further analsys")
						fmt.Println()
					}
				}
			}

		case <-time.After(25 * time.Second):
			if !isPodReady {
				fmt.Printf("error: %s pod was not ready in time", mizu.AggregatorPodName)
				cancel()
			}

		case <-errorChan:
			cancel()
		}
	}
}

func createRBACIfNecessary(ctx context.Context, kubernetesProvider *kubernetes.Provider) bool {
	mizuRBACExists, err := kubernetesProvider.DoesServiceAccountExist(ctx, mizu.ResourcesNamespace, mizu.ServiceAccountName)
	if err != nil {
		fmt.Printf("warning: could not ensure mizu rbac resources exist %v\n", err)
		return false
	}
	if !mizuRBACExists {
		err := kubernetesProvider.CreateMizuRBAC(ctx, mizu.ResourcesNamespace, mizu.ServiceAccountName, mizu.ClusterRoleName, mizu.ClusterRoleBindingName, mizu.RBACVersion)
		if err != nil {
			fmt.Printf("warning: could not create mizu rbac resources %v\n", err)
			return false
		}
	}
	return true
}

func getNodeHostToTappedPodIpsMap(tappedPods []core.Pod) (map[string][]string, error) {
	nodeToTappedPodIPMap := make(map[string][]string, 0)
	for _, pod := range tappedPods {
		existingList := nodeToTappedPodIPMap[pod.Spec.NodeName]
		if existingList == nil {
			nodeToTappedPodIPMap[pod.Spec.NodeName] = []string{pod.Status.PodIP}
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
	case <-ctx.Done():
		break
	case <-sigChan:
		cancel()
	}
}

func syncApiStatus(ctx context.Context, cancel context.CancelFunc, tappingOptions *MizuTapOptions) {
	controlSocketStr := fmt.Sprintf("ws://%s/ws", kubernetes.GetMizuCollectorProxiedHostAndPath(tappingOptions.GuiPort, mizu.ResourcesNamespace, mizu.AggregatorPodName))
	controlSocket, err := mizu.CreateControlSocket(controlSocketStr)
	if err != nil {
		fmt.Printf("error establishing control socket connection %s\n", err)
		cancel()
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			err = controlSocket.SendNewTappedPodsListMessage(currentlyTappedPods)
			if err != nil {
				rlog.Debugf("error Sending message via control socket %v, error: %s\n", controlSocketStr, err)
			}
			time.Sleep(10 * time.Second)
		}
	}

}

func getNamespace(tappingOptions *MizuTapOptions, kubernetesProvider *kubernetes.Provider) string {
	if tappingOptions.AllNamespaces {
		return mizu.K8sAllNamespaces
	} else if len(tappingOptions.Namespace) > 0 {
		return tappingOptions.Namespace
	} else {
		return kubernetesProvider.CurrentNamespace()
	}
}
