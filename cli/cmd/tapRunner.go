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
	var namespacesStr string
	if targetNamespace != mizu.K8sAllNamespaces {
		namespacesStr = fmt.Sprintf("namespace \"%s\"", targetNamespace)
	} else {
		namespacesStr = "all namespaces"
	}
	mizu.CheckNewerVersion()
	mizu.Log.Infof("Tapping pods in %s", namespacesStr)

	if err, _ := updateCurrentlyTappedPods(kubernetesProvider, ctx, targetNamespace); err != nil {
		mizu.Log.Infof("Error listing pods: %v", err)
		return
	}

	if len(currentlyTappedPods) == 0 {
		var suggestionStr string
		if targetNamespace != mizu.K8sAllNamespaces {
			suggestionStr = "\nSelect a different namespace with -n or tap all namespaces with -A"
		}
		mizu.Log.Infof("Did not find any pods matching the regex argument%s", suggestionStr)
		return
	}

	if mizu.Config.Tap.DryRun {
		return
	}

	nodeToTappedPodIPMap, err := getNodeHostToTappedPodIpsMap(currentlyTappedPods)
	if err != nil {
		return
	}

	if err := createMizuResources(ctx, kubernetesProvider, nodeToTappedPodIPMap, mizuApiFilteringOptions); err != nil {
		return
	}

	go createProxyToApiServerPod(ctx, kubernetesProvider, cancel)
	go watchPodsForTapping(ctx, kubernetesProvider, cancel)

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
		return err
	}
	mizu.Log.Debugf("Successfully creating Namespace %s", mizu.ResourcesNamespace)
	return nil
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
	mizu.Log.Debugf("Successfully created API server pod: %s", mizu.ApiServerPodName)

	apiServerService, err = kubernetesProvider.CreateService(ctx, mizu.ResourcesNamespace, mizu.ApiServerPodName, mizu.ApiServerPodName)
	if err != nil {
		mizu.Log.Infof("Error creating mizu %s service: %v", mizu.ApiServerPodName, err)
		return err
	}
	mizu.Log.Debugf("Successfully created service: %s", mizu.ApiServerPodName)

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
		mizu.Log.Debugf("Successfully created %v tappers", len(nodeToTappedPodIPMap))
	} else {
		if err := kubernetesProvider.RemoveDaemonSet(ctx, mizu.ResourcesNamespace, mizu.TapperDaemonSetName); err != nil {
			mizu.Log.Errorf("Error deleting mizu tapper daemonset: %v", err)
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

	controlSocketStr := fmt.Sprintf("ws://%s/ws", kubernetes.GetMizuApiServerProxiedHostAndPath(mizu.Config.Tap.GuiPort))
	controlSocket, err := mizu.CreateControlSocket(controlSocketStr)
	if err != nil {
		mizu.Log.Infof("error establishing control socket connection %s", err)
		cancel()
	}
	mizu.Log.Debugf("Control socket created %s", controlSocketStr)
	err = controlSocket.SendNewTappedPodsListMessage(currentlyTappedPods)
	if err != nil {
		mizu.Log.Debugf("error Sending message via control socket %v, error: %s", controlSocketStr, err)
	}
	restartTappers := func() {
		err, changeFound := updateCurrentlyTappedPods(kubernetesProvider, ctx, targetNamespace)
		if err != nil {
			mizu.Log.Errorf("Error getting pods by regex: %s (%v,%+v)", err, err, err)
			cancel()
		}

		if !changeFound {
			mizu.Log.Debugf("Nothing changed update tappers not needed")
			return
		}

		err = controlSocket.SendNewTappedPodsListMessage(currentlyTappedPods)
		if err != nil {
			mizu.Log.Debugf("error Sending message via control socket %v, error: %s", controlSocketStr, err)
		}

		nodeToTappedPodIPMap, err := getNodeHostToTappedPodIpsMap(currentlyTappedPods)
		if err != nil {
			mizu.Log.Errorf("Error building node to ips map: %s (%v,%+v)", err, err, err)
			cancel()
		}

		if err := updateMizuTappers(ctx, kubernetesProvider, nodeToTappedPodIPMap); err != nil {
			mizu.Log.Errorf("Error updating daemonset: %s (%v,%+v)", err, err, err)
			cancel()
		}
	}
	restartTappersDebouncer := debounce.NewDebouncer(updateTappersDelay, restartTappers)

	for {
		select {
		case pod := <-added:
			mizu.Log.Debugf("Added matching pod %s, ns: %s", pod.Name, pod.Namespace)
			restartTappersDebouncer.SetOn()
		case pod := <-removed:
			mizu.Log.Debugf("Removed matching pod %s, ns: %s", pod.Name, pod.Namespace)
			restartTappersDebouncer.SetOn()
		case pod := <-modified:
			mizu.Log.Debugf("Modified matching pod %s, ns: %s, phase: %s, ip: %s", pod.Name, pod.Namespace, pod.Status.Phase, pod.Status.PodIP)
			// Act only if the modified pod has already obtained an IP address.
			// After filtering for IPs, on a normal pod restart this includes the following events:
			// - Pod deletion
			// - Pod reaches start state
			// - Pod reaches ready state
			// Ready/unready transitions might also trigger this event.
			if pod.Status.PodIP != "" {
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

func updateCurrentlyTappedPods(kubernetesProvider *kubernetes.Provider, ctx context.Context, targetNamespace string) (error, bool) {
	changeFound := false
	if matchingPods, err := kubernetesProvider.GetAllRunningPodsMatchingRegex(ctx, mizu.Config.Tap.PodRegex(), targetNamespace); err != nil {
		mizu.Log.Infof("Error getting pods by regex: %s (%v,%+v)", err, err, err)
		return err, false
	} else {
		addedPods, removedPods := getPodArrayDiff(currentlyTappedPods, matchingPods)
		for _, addedPod := range addedPods {
			changeFound = true
			mizu.Log.Infof(uiUtils.Green, fmt.Sprintf("+%s", addedPod.Name))
		}
		for _, removedPod := range removedPods {
			changeFound = true
			mizu.Log.Infof(uiUtils.Red, fmt.Sprintf("-%s", removedPod.Name))
		}
		currentlyTappedPods = matchingPods
	}

	return nil, changeFound
}

func getPodArrayDiff(oldPods []core.Pod, newPods []core.Pod) (added []core.Pod, removed []core.Pod) {
	added = getMissingPods(newPods, oldPods)
	removed = getMissingPods(oldPods, newPods)

	return added, removed
}

//returns pods present in pods1 array and missing in pods2 array
func getMissingPods(pods1 []core.Pod, pods2 []core.Pod) []core.Pod {
	missingPods := make([]core.Pod, 0)
	for _, pod1 := range pods1 {
		var found = false
		for _, pod2 := range pods2 {
			if pod1.UID == pod2.UID {
				found = true
				break
			}
		}
		if !found {
			missingPods = append(missingPods, pod1)
		}
	}
	return missingPods
}

func createProxyToApiServerPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc) {
	podExactRegex := regexp.MustCompile(fmt.Sprintf("^%s$", mizu.ApiServerPodName))
	added, modified, removed, errorChan := kubernetes.FilteredWatch(ctx, kubernetesProvider.GetPodWatcher(ctx, mizu.ResourcesNamespace), podExactRegex)
	isPodReady := false
	timeAfter := time.After(25 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-added:
			mizu.Log.Debugf("Got agent pod added event")
			continue
		case <-removed:
			mizu.Log.Infof("%s removed", mizu.ApiServerPodName)
			cancel()
			return
		case modifiedPod := <-modified:
			mizu.Log.Debugf("Got agent pod modified event, status phase: %v", modifiedPod.Status.Phase)
			if modifiedPod.Status.Phase == core.PodRunning && !isPodReady {
				isPodReady = true
				go func() {
					err := kubernetes.StartProxy(kubernetesProvider, mizu.Config.Tap.GuiPort, mizu.ResourcesNamespace, mizu.ApiServerPodName)
					if err != nil {
						mizu.Log.Errorf("Error occurred while running k8s proxy %v", err)
						cancel()
					}
				}()
				mizu.Log.Infof("Mizu is available at http://%s\n", kubernetes.GetMizuApiServerProxiedHostAndPath(mizu.Config.Tap.GuiPort))
				time.Sleep(time.Second * 5) // Waiting to be sure the proxy is ready
				requestForAnalysis()
			}
		case <-timeAfter:
			if !isPodReady {
				mizu.Log.Errorf("Error: %s pod was not ready in time", mizu.ApiServerPodName)
				cancel()
			}
		case <-errorChan:
			mizu.Log.Debugf("[ERROR] Agent creation, watching %v namespace", mizu.ResourcesNamespace)
			cancel()
		}
	}
}

func requestForAnalysis() {
	if !mizu.Config.Tap.Analysis {
		return
	}

	mizuProxiedUrl := kubernetes.GetMizuApiServerProxiedHostAndPath(mizu.Config.Tap.GuiPort)
	urlPath := fmt.Sprintf("http://%s/api/uploadEntries?dest=%s&interval=%v", mizuProxiedUrl, url.QueryEscape(mizu.Config.Tap.AnalysisDestination), mizu.Config.Tap.SleepIntervalSec)
	u, parseErr := url.ParseRequestURI(urlPath)
	if parseErr != nil {
		mizu.Log.Fatal("Failed parsing the URL (consider changing the analysis dest URL), err: %v", parseErr)
	}

	mizu.Log.Debugf("Sending get request to %v", u.String())
	if response, requestErr := http.Get(u.String()); requestErr != nil {
		mizu.Log.Errorf("Failed to notify agent for analysis, err: %v", requestErr)
	} else if response.StatusCode != 200 {
		mizu.Log.Errorf("Failed to notify agent for analysis, status code: %v", response.StatusCode)
	} else {
		mizu.Log.Infof(uiUtils.Purple, "Traffic is uploading to UP9 for further analysis")
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

func getNamespace(kubernetesProvider *kubernetes.Provider) string {
	if mizu.Config.Tap.AllNamespaces {
		return mizu.K8sAllNamespaces
	} else if len(mizu.Config.Tap.Namespace) > 0 {
		return mizu.Config.Tap.Namespace
	} else {
		return kubernetesProvider.CurrentNamespace()
	}
}
