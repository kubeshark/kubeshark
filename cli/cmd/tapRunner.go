package cmd

import (
	"context"
	"fmt"
	"github.com/up9inc/mizu/cli/kubernetes"
	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/debounce"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"
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
var apiServerService *core.Service

const (
	updateTappersDelay = 5 * time.Second
	cleanupTimeout     = time.Minute
)

var currentlyTappedPods []core.Pod

func RunMizuTap() {
	mizuApiFilteringOptions, err := getMizuApiFilteringOptions()
	if err != nil {
		return
	}

	kubernetesProvider, err := kubernetes.NewProvider(mizu.Config.Tap.KubeConfigPath)
	if err != nil {
		if clientcmd.IsEmptyConfig(err) {
			mizu.Log.Infof(uiUtils.Red, "Couldn't find the kube config file, or file is empty. Try adding '--kube-config=<path to kube config file>'\n")
			return
		}
		if clientcmd.IsConfigurationInvalid(err) {
			mizu.Log.Infof(uiUtils.Red, "Invalid kube config file. Try using a different config with '--kube-config=<path to kube config file>'\n")
			return
		}
	}

	defer cleanUpMizuResources(kubernetesProvider)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	targetNamespace := getNamespace(kubernetesProvider)
	if matchingPods, err := kubernetesProvider.GetAllPodsMatchingRegex(ctx, mizu.Config.Tap.PodRegex(), targetNamespace); err != nil {
		mizu.Log.Infof("Error listing pods: %v", err)
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
	mizu.Log.Infof("Tapping pods in %s", namespacesStr)

	if len(currentlyTappedPods) == 0 {
		var suggestionStr string
		if targetNamespace != mizu.K8sAllNamespaces {
			suggestionStr = "\nSelect a different namespace with -n or tap all namespaces with -A"
		}
		mizu.Log.Infof("Did not find any pods matching the regex argument%s", suggestionStr)
	}

	if mizu.Config.Tap.TappedPodsPreview {
		for _, tappedPod := range currentlyTappedPods {
			mizu.Log.Infof(tappedPod.Name)
		}
		return
	}

	nodeToTappedPodIPMap, err := getNodeHostToTappedPodIpsMap(currentlyTappedPods)
	if err != nil {
		return
	}

	if err := createMizuResources(ctx, kubernetesProvider, nodeToTappedPodIPMap, mizuApiFilteringOptions); err != nil {
		return
	}

	mizu.CheckNewerVersion()
	go portForwardApiPod(ctx, kubernetesProvider, cancel) // TODO convert this to job for built in pod ttl or have the running app handle this
	go watchPodsForTapping(ctx, kubernetesProvider, cancel)
	go syncApiStatus(ctx, cancel)

	//block until exit signal or error
	waitForFinish(ctx, cancel)
}

func createMizuResources(ctx context.Context, kubernetesProvider *kubernetes.Provider, nodeToTappedPodIPMap map[string][]string, mizuApiFilteringOptions *shared.TrafficFilteringOptions) error {
	if err := createMizuNamespace(ctx, kubernetesProvider); err != nil {
		return err
	}

	if err := createMizuApiServer(ctx, kubernetesProvider, mizuApiFilteringOptions); err != nil {
		return err
	}

	if err := updateMizuTappers(ctx, kubernetesProvider, nodeToTappedPodIPMap); err != nil {
		return err
	}

	return nil
}

func createMizuNamespace(ctx context.Context, kubernetesProvider *kubernetes.Provider) error {
	_, err := kubernetesProvider.CreateNamespace(ctx, mizu.ResourcesNamespace)
	if err != nil {
		mizu.Log.Infof("Error creating Namespace %s: %v", mizu.ResourcesNamespace, err)
	}

	return err
}

func createMizuApiServer(ctx context.Context, kubernetesProvider *kubernetes.Provider, mizuApiFilteringOptions *shared.TrafficFilteringOptions) error {
	var err error

	mizuServiceAccountExists = createRBACIfNecessary(ctx, kubernetesProvider)
	var serviceAccountName string
	if mizuServiceAccountExists {
		serviceAccountName = mizu.ServiceAccountName
	} else {
		serviceAccountName = ""
	}
	_, err = kubernetesProvider.CreateMizuApiServerPod(ctx, mizu.ResourcesNamespace, mizu.ApiServerPodName, mizu.Config.MizuImage, serviceAccountName, mizuApiFilteringOptions, mizu.Config.Tap.MaxEntriesDBSizeBytes())
	if err != nil {
		mizu.Log.Infof("Error creating mizu %s pod: %v", mizu.ApiServerPodName, err)
		return err
	}

	apiServerService, err = kubernetesProvider.CreateService(ctx, mizu.ResourcesNamespace, mizu.ApiServerPodName, mizu.ApiServerPodName)
	if err != nil {
		mizu.Log.Infof("Error creating mizu %s service: %v", mizu.ApiServerPodName, err)
		return err
	}

	return nil
}

func getMizuApiFilteringOptions() (*shared.TrafficFilteringOptions, error) {
	var compiledRegexSlice []*shared.SerializableRegexp

	if mizu.Config.Tap.PlainTextFilterRegexes != nil && len(mizu.Config.Tap.PlainTextFilterRegexes) > 0 {
		compiledRegexSlice = make([]*shared.SerializableRegexp, 0)
		for _, regexStr := range mizu.Config.Tap.PlainTextFilterRegexes {
			compiledRegex, err := shared.CompileRegexToSerializableRegexp(regexStr)
			if err != nil {
				mizu.Log.Infof("Regex %s is invalid: %v", regexStr, err)
				return nil, err
			}
			compiledRegexSlice = append(compiledRegexSlice, compiledRegex)
		}
	}

	return &shared.TrafficFilteringOptions{PlainTextMaskingRegexes: compiledRegexSlice, HideHealthChecks: mizu.Config.Tap.HideHealthChecks, DisableRedaction: mizu.Config.Tap.DisableRedaction}, nil
}

func updateMizuTappers(ctx context.Context, kubernetesProvider *kubernetes.Provider, nodeToTappedPodIPMap map[string][]string) error {
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
			mizu.Config.MizuImage,
			mizu.TapperPodName,
			fmt.Sprintf("%s.%s.svc.cluster.local", apiServerService.Name, apiServerService.Namespace),
			nodeToTappedPodIPMap,
			serviceAccountName,
			mizu.Config.Tap.TapOutgoing(),
		); err != nil {
			mizu.Log.Infof("Error creating mizu tapper daemonset: %v", err)
			return err
		}
	} else {
		if err := kubernetesProvider.RemoveDaemonSet(ctx, mizu.ResourcesNamespace, mizu.TapperDaemonSetName); err != nil {
			mizu.Log.Infof("Error deleting mizu tapper daemonset: %v", err)
			return err
		}
	}

	return nil
}

func cleanUpMizuResources(kubernetesProvider *kubernetes.Provider) {
	mizu.Log.Infof("\nRemoving mizu resources\n")

	removalCtx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()

	if err := kubernetesProvider.RemoveNamespace(removalCtx, mizu.ResourcesNamespace); err != nil {
		mizu.Log.Infof("Error removing Namespace %s: %s (%v,%+v)", mizu.ResourcesNamespace, err, err, err)
		return
	}

	if mizuServiceAccountExists {
		if err := kubernetesProvider.RemoveNonNamespacedResources(removalCtx, mizu.ClusterRoleName, mizu.ClusterRoleBindingName); err != nil {
			mizu.Log.Infof("Error removing non-namespaced resources: %s (%v,%+v)", err, err, err)
			return
		}
	}

	// Call cancel if a terminating signal was received. Allows user to skip the wait.
	go func() {
		waitForFinish(removalCtx, cancel)
	}()

	if err := kubernetesProvider.WaitUtilNamespaceDeleted(removalCtx, mizu.ResourcesNamespace); err != nil {
		switch {
		case removalCtx.Err() == context.Canceled:
			// Do nothing. User interrupted the wait.
		case err == wait.ErrWaitTimeout:
			mizu.Log.Infof("Timeout while removing Namespace %s", mizu.ResourcesNamespace)
		default:
			mizu.Log.Infof("Error while waiting for Namespace %s to be deleted: %s (%v,%+v)", mizu.ResourcesNamespace, err, err, err)
		}
	}
}

func watchPodsForTapping(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc) {
	targetNamespace := getNamespace(kubernetesProvider)

	added, modified, removed, errorChan := kubernetes.FilteredWatch(ctx, kubernetesProvider.GetPodWatcher(ctx, targetNamespace), mizu.Config.Tap.PodRegex())

	restartTappers := func() {
		if matchingPods, err := kubernetesProvider.GetAllPodsMatchingRegex(ctx, mizu.Config.Tap.PodRegex(), targetNamespace); err != nil {
			mizu.Log.Infof("Error getting pods by regex: %s (%v,%+v)", err, err, err)
			cancel()
		} else {
			currentlyTappedPods = matchingPods
		}

		nodeToTappedPodIPMap, err := getNodeHostToTappedPodIpsMap(currentlyTappedPods)
		if err != nil {
			mizu.Log.Infof("Error building node to ips map: %s (%v,%+v)", err, err, err)
			cancel()
		}

		if err := updateMizuTappers(ctx, kubernetesProvider, nodeToTappedPodIPMap); err != nil {
			mizu.Log.Infof("Error updating daemonset: %s (%v,%+v)", err, err, err)
			cancel()
		}
	}
	restartTappersDebouncer := debounce.NewDebouncer(updateTappersDelay, restartTappers)

	for {
		select {
		case newTarget := <-added:
			mizu.Log.Infof(uiUtils.Green, fmt.Sprintf("+%s", newTarget.Name))

		case removedTarget := <-removed:
			mizu.Log.Infof(uiUtils.Red, fmt.Sprintf("-%s", removedTarget.Name))
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

func portForwardApiPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc) {
	podExactRegex := regexp.MustCompile(fmt.Sprintf("^%s$", mizu.ApiServerPodName))
	added, modified, removed, errorChan := kubernetes.FilteredWatch(ctx, kubernetesProvider.GetPodWatcher(ctx, mizu.ResourcesNamespace), podExactRegex)
	isPodReady := false
	timeAfter := time.After(25 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return

		case <-added:
			continue
		case <-removed:
			mizu.Log.Infof("%s removed", mizu.ApiServerPodName)
			cancel()
			return
		case modifiedPod := <-modified:
			if modifiedPod.Status.Phase == "Running" && !isPodReady {
				isPodReady = true
				go func() {
					err := kubernetes.StartProxy(kubernetesProvider, mizu.Config.Tap.GuiPort, mizu.ResourcesNamespace, mizu.ApiServerPodName)
					if err != nil {
						mizu.Log.Infof("Error occured while running k8s proxy %v", err)
						cancel()
					}
				}()
				mizuProxiedUrl := kubernetes.GetMizuApiServerProxiedHostAndPath(mizu.Config.Tap.GuiPort)
				mizu.Log.Infof("Mizu is available at http://%s", mizuProxiedUrl)

				time.Sleep(time.Second * 5) // Waiting to be sure the proxy is ready
				if mizu.Config.Tap.Analysis {
					urlPath := fmt.Sprintf("http://%s/api/uploadEntries?dest=%s&interval=%v", mizuProxiedUrl, url.QueryEscape(mizu.Config.Tap.AnalysisDestination), mizu.Config.Tap.SleepIntervalSec)
					u, err := url.ParseRequestURI(urlPath)

					if err != nil {
						log.Fatal(fmt.Sprintf("Failed parsing the URL %v\n", err))
					}
					mizu.Log.Debugf("Sending get request to %v", u.String())
					if response, err := http.Get(u.String()); err != nil || response.StatusCode != 200 {
						mizu.Log.Infof("error sending upload entries req, status code: %v, err: %v", response.StatusCode, err)
					} else {
						mizu.Log.Infof(uiUtils.Purple, "Traffic is uploading to UP9 for further analysis")
					}
				}
			}

		case <-timeAfter:
			if !isPodReady {
				mizu.Log.Infof("error: %s pod was not ready in time", mizu.ApiServerPodName)
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
		mizu.Log.Infof("warning: could not ensure mizu rbac resources exist %v", err)
		return false
	}
	if !mizuRBACExists {
		err := kubernetesProvider.CreateMizuRBAC(ctx, mizu.ResourcesNamespace, mizu.ServiceAccountName, mizu.ClusterRoleName, mizu.ClusterRoleBindingName, mizu.RBACVersion)
		if err != nil {
			mizu.Log.Infof("warning: could not create mizu rbac resources %v", err)
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

func syncApiStatus(ctx context.Context, cancel context.CancelFunc) {
	controlSocketStr := fmt.Sprintf("ws://%s/ws", kubernetes.GetMizuApiServerProxiedHostAndPath(mizu.Config.Tap.GuiPort))
	controlSocket, err := mizu.CreateControlSocket(controlSocketStr)
	if err != nil {
		mizu.Log.Infof("error establishing control socket connection %s", err)
		cancel()
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			err = controlSocket.SendNewTappedPodsListMessage(currentlyTappedPods)
			if err != nil {
				mizu.Log.Debugf("error Sending message via control socket %v, error: %s", controlSocketStr, err)
			}
			time.Sleep(10 * time.Second)
		}
	}

}

func getNamespace(kubernetesProvider *kubernetes.Provider) string {
	if mizu.Config.Tap.AllNamespaces {
		return mizu.K8sAllNamespaces
	} else if len(mizu.Config.Tap.Namespace) > 0 {
		return mizu.Config.Tap.Namespace
	} else {
		return kubernetesProvider.CurrentNamespace()
	}
}
