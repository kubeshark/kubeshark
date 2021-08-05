package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/up9inc/mizu/cli/errormessage"
	"github.com/up9inc/mizu/cli/kubernetes"
	"github.com/up9inc/mizu/cli/logsUtils"
	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/debounce"
	yaml "gopkg.in/yaml.v3"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	cleanupTimeout     = time.Minute
	updateTappersDelay = 5 * time.Second
)

type tapState struct {
	apiServerService         *core.Service
	currentlyTappedPods      []core.Pod
	mizuServiceAccountExists bool
	doNotRemoveConfigMap     bool
}

var state tapState

func RunMizuTap() {
	mizuApiFilteringOptions, err := getMizuApiFilteringOptions()
	if err != nil {
		mizu.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error parsing regex-masking: %v", errormessage.FormatError(err)))
		return
	}
	var mizuValidationRules string
	if mizu.Config.Tap.EnforcePolicyFile != "" {
		mizuValidationRules, err = readValidationRules(mizu.Config.Tap.EnforcePolicyFile)
		if err != nil {
			mizu.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error reading policy file: %v", errormessage.FormatError(err)))
			return
		}
	}

	kubernetesProvider, err := kubernetes.NewProvider(mizu.Config.Tap.KubeConfigPath)
	if err != nil {
		mizu.Log.Error(err)
		return
	}

	defer cleanUpMizuResources(kubernetesProvider)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	targetNamespaces := getNamespaces(kubernetesProvider)

	var namespacesStr string
	if targetNamespaces[0] != mizu.K8sAllNamespaces {
		namespacesStr = fmt.Sprintf("namespaces \"%s\"", strings.Join(targetNamespaces, "\", \""))
	} else {
		namespacesStr = "all namespaces"
	}
	mizu.CheckNewerVersion()
	mizu.Log.Infof("Tapping pods in %s", namespacesStr)

	if err, _ := updateCurrentlyTappedPods(kubernetesProvider, ctx, targetNamespaces); err != nil {
		mizu.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error getting pods by regex: %v", errormessage.FormatError(err)))
		return
	}

	if len(state.currentlyTappedPods) == 0 {
		var suggestionStr string
		if targetNamespaces[0] != mizu.K8sAllNamespaces {
			suggestionStr = ". Select a different namespace with -n or tap all namespaces with -A"
		}
		mizu.Log.Warningf(uiUtils.Warning, fmt.Sprintf("Did not find any pods matching the regex argument%s", suggestionStr))
	}

	if mizu.Config.Tap.DryRun {
		return
	}

	nodeToTappedPodIPMap := getNodeHostToTappedPodIpsMap(state.currentlyTappedPods)

	if err := createMizuResources(ctx, kubernetesProvider, nodeToTappedPodIPMap, mizuApiFilteringOptions, mizuValidationRules); err != nil {
		mizu.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error creating resources: %v", errormessage.FormatError(err)))
		return
	}

	go createProxyToApiServerPod(ctx, kubernetesProvider, cancel)
	go watchPodsForTapping(ctx, kubernetesProvider, targetNamespaces, cancel)

	//block until exit signal or error
	waitForFinish(ctx, cancel)
}

func readValidationRules(file string) (string, error) {
	rules, err := shared.DecodeEnforcePolicy(file)
	if err != nil {
		return "", err
	}
	newContent, _ := yaml.Marshal(&rules)
	return string(newContent), nil
}

func createMizuResources(ctx context.Context, kubernetesProvider *kubernetes.Provider, nodeToTappedPodIPMap map[string][]string, mizuApiFilteringOptions *shared.TrafficFilteringOptions, mizuValidationRules string) error {
	if !mizu.Config.IsNsRestrictedMode() {
		if err := createMizuNamespace(ctx, kubernetesProvider); err != nil {
			return err
		}
	}

	if err := createMizuApiServer(ctx, kubernetesProvider, mizuApiFilteringOptions); err != nil {
		return err
	}

	if err := updateMizuTappers(ctx, kubernetesProvider, nodeToTappedPodIPMap); err != nil {
		return err
	}

	if err := createMizuConfigmap(ctx, kubernetesProvider, mizuValidationRules); err != nil {
		mizu.Log.Warningf(uiUtils.Warning, fmt.Sprintf("Failed to create resources required for policy validation. Mizu will not validate policy rules. error: %v\n", errormessage.FormatError(err)))
		state.doNotRemoveConfigMap = true
	} else if mizuValidationRules == "" {
		state.doNotRemoveConfigMap = true
	}

	return nil
}

func createMizuConfigmap(ctx context.Context, kubernetesProvider *kubernetes.Provider, data string) error {
	err := kubernetesProvider.CreateConfigMap(ctx, mizu.Config.MizuResourcesNamespace, mizu.ConfigMapName, data)
	return err
}

func createMizuNamespace(ctx context.Context, kubernetesProvider *kubernetes.Provider) error {
	_, err := kubernetesProvider.CreateNamespace(ctx, mizu.Config.MizuResourcesNamespace)
	return err
}

func createMizuApiServer(ctx context.Context, kubernetesProvider *kubernetes.Provider, mizuApiFilteringOptions *shared.TrafficFilteringOptions) error {
	var err error

	state.mizuServiceAccountExists, err = createRBACIfNecessary(ctx, kubernetesProvider)
	if err != nil {
		mizu.Log.Warningf(uiUtils.Warning, fmt.Sprintf("Failed to ensure the resources required for IP resolving. Mizu will not resolve target IPs to names. error: %v", errormessage.FormatError(err)))
	}

	var serviceAccountName string
	if state.mizuServiceAccountExists {
		serviceAccountName = mizu.ServiceAccountName
	} else {
		serviceAccountName = ""
	}

	opts := &kubernetes.ApiServerOptions{
		Namespace:               mizu.Config.MizuResourcesNamespace,
		PodName:                 mizu.ApiServerPodName,
		PodImage:                mizu.Config.AgentImage,
		ServiceAccountName:      serviceAccountName,
		IsNamespaceRestricted:   mizu.Config.IsNsRestrictedMode(),
		MizuApiFilteringOptions: mizuApiFilteringOptions,
		MaxEntriesDBSizeBytes:   mizu.Config.Tap.MaxEntriesDBSizeBytes(),
	}
	_, err = kubernetesProvider.CreateMizuApiServerPod(ctx, opts)
	if err != nil {
		return err
	}
	mizu.Log.Debugf("Successfully created API server pod: %s", mizu.ApiServerPodName)

	state.apiServerService, err = kubernetesProvider.CreateService(ctx, mizu.Config.MizuResourcesNamespace, mizu.ApiServerPodName, mizu.ApiServerPodName)
	if err != nil {
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
		if state.mizuServiceAccountExists {
			serviceAccountName = mizu.ServiceAccountName
		} else {
			serviceAccountName = ""
		}

		if err := kubernetesProvider.ApplyMizuTapperDaemonSet(
			ctx,
			mizu.Config.MizuResourcesNamespace,
			mizu.TapperDaemonSetName,
			mizu.Config.AgentImage,
			mizu.TapperPodName,
			fmt.Sprintf("%s.%s.svc.cluster.local", state.apiServerService.Name, state.apiServerService.Namespace),
			nodeToTappedPodIPMap,
			serviceAccountName,
			mizu.Config.Tap.TapOutgoing(),
		); err != nil {
			return err
		}
		mizu.Log.Debugf("Successfully created %v tappers", len(nodeToTappedPodIPMap))
	} else {
		if err := kubernetesProvider.RemoveDaemonSet(ctx, mizu.Config.MizuResourcesNamespace, mizu.TapperDaemonSetName); err != nil {
			return err
		}
	}

	return nil
}

func cleanUpMizuResources(kubernetesProvider *kubernetes.Provider) {

	removalCtx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()

	if mizu.Config.DumpLogs {
		mizuDir := mizu.GetMizuFolderPath()
		filePath = path.Join(mizuDir, fmt.Sprintf("mizu_logs_%s.zip", time.Now().Format("2006_01_02__15_04_05")))
		if err := logsUtils.DumpLogs(kubernetesProvider, removalCtx, filePath); err != nil {
			mizu.Log.Errorf("Failed dump logs %v", err)
		}
	}

	mizu.Log.Infof("\nRemoving mizu resources\n")

	if !mizu.Config.IsNsRestrictedMode() {
		if err := kubernetesProvider.RemoveNamespace(removalCtx, mizu.Config.MizuResourcesNamespace); err != nil {
			mizu.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error removing Namespace %s: %v", mizu.Config.MizuResourcesNamespace, errormessage.FormatError(err)))
			return
		}
	} else {
		if err := kubernetesProvider.RemovePod(removalCtx, mizu.Config.MizuResourcesNamespace, mizu.ApiServerPodName); err != nil {
			mizu.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error removing Pod %s in namespace %s: %v", mizu.ApiServerPodName, mizu.Config.MizuResourcesNamespace, errormessage.FormatError(err)))
		}

		if err := kubernetesProvider.RemoveService(removalCtx, mizu.Config.MizuResourcesNamespace, mizu.ApiServerPodName); err != nil {
			mizu.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error removing Service %s in namespace %s: %v", mizu.ApiServerPodName, mizu.Config.MizuResourcesNamespace, errormessage.FormatError(err)))
		}

		if err := kubernetesProvider.RemoveDaemonSet(removalCtx, mizu.Config.MizuResourcesNamespace, mizu.TapperDaemonSetName); err != nil {
			mizu.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error removing DaemonSet %s in namespace %s: %v", mizu.TapperDaemonSetName, mizu.Config.MizuResourcesNamespace, errormessage.FormatError(err)))
		}

		if !state.doNotRemoveConfigMap {
			if err := kubernetesProvider.RemoveConfigMap(removalCtx, mizu.Config.MizuResourcesNamespace, mizu.ConfigMapName); err != nil {
				mizu.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error removing ConfigMap %s in namespace %s: %v", mizu.ConfigMapName, mizu.Config.MizuResourcesNamespace, errormessage.FormatError(err)))
			}
		}

	}

	if state.mizuServiceAccountExists {
		if !mizu.Config.IsNsRestrictedMode() {
			if err := kubernetesProvider.RemoveNonNamespacedResources(removalCtx, mizu.ClusterRoleName, mizu.ClusterRoleBindingName); err != nil {
				mizu.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error removing non-namespaced resources: %v", errormessage.FormatError(err)))
				return
			}
		} else {
			if err := kubernetesProvider.RemoveServicAccount(removalCtx, mizu.Config.MizuResourcesNamespace, mizu.ServiceAccountName); err != nil {
				mizu.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error removing Service Account %s in namespace %s: %v", mizu.ServiceAccountName, mizu.Config.MizuResourcesNamespace, errormessage.FormatError(err)))
				return
			}

			if err := kubernetesProvider.RemoveRole(removalCtx, mizu.Config.MizuResourcesNamespace, mizu.RoleName); err != nil {
				mizu.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error removing Role %s in namespace %s: %v", mizu.RoleName, mizu.Config.MizuResourcesNamespace, errormessage.FormatError(err)))
			}

			if err := kubernetesProvider.RemoveRoleBinding(removalCtx, mizu.Config.MizuResourcesNamespace, mizu.RoleBindingName); err != nil {
				mizu.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error removing RoleBinding %s in namespace %s: %v", mizu.RoleBindingName, mizu.Config.MizuResourcesNamespace, errormessage.FormatError(err)))
			}
		}
	}

	if !mizu.Config.IsNsRestrictedMode() {
		waitUntilNamespaceDeleted(removalCtx, cancel, kubernetesProvider)
	}
}

func waitUntilNamespaceDeleted(ctx context.Context, cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider) {
	// Call cancel if a terminating signal was received. Allows user to skip the wait.
	go func() {
		waitForFinish(ctx, cancel)
	}()

	if err := kubernetesProvider.WaitUtilNamespaceDeleted(ctx, mizu.Config.MizuResourcesNamespace); err != nil {
		switch {
		case ctx.Err() == context.Canceled:
			// Do nothing. User interrupted the wait.
		case err == wait.ErrWaitTimeout:
			mizu.Log.Errorf(uiUtils.Error, fmt.Sprintf("Timeout while removing Namespace %s", mizu.Config.MizuResourcesNamespace))
		default:
			mizu.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error while waiting for Namespace %s to be deleted: %v", mizu.Config.MizuResourcesNamespace, errormessage.FormatError(err)))
		}
	}
}

func reportTappedPods() {
	mizuProxiedUrl := kubernetes.GetMizuApiServerProxiedHostAndPath(mizu.Config.Fetch.MizuPort)
	tappedPodsUrl := fmt.Sprintf("http://%s/status/tappedPods", mizuProxiedUrl)

	podInfos := make([]shared.PodInfo, 0)
	for _, pod := range state.currentlyTappedPods {
		podInfos = append(podInfos, shared.PodInfo{Name: pod.Name, Namespace: pod.Namespace})
	}
	tapStatus := shared.TapStatus{Pods: podInfos}

	if jsonValue, err := json.Marshal(tapStatus); err != nil {
		mizu.Log.Debugf("[ERROR] failed Marshal the tapped pods %v", err)
	} else {
		if response, err := http.Post(tappedPodsUrl, "application/json", bytes.NewBuffer(jsonValue)); err != nil {
			mizu.Log.Debugf("[ERROR] failed sending to API server the tapped pods %v", err)
		} else if response.StatusCode != 200 {
			mizu.Log.Debugf("[ERROR] failed sending to API server the tapped pods, response status code %v", response.StatusCode)
		} else {
			mizu.Log.Debugf("Reported to server API about %d taped pods successfully", len(podInfos))
		}
	}
}

func watchPodsForTapping(ctx context.Context, kubernetesProvider *kubernetes.Provider, targetNamespaces []string, cancel context.CancelFunc) {
	added, modified, removed, errorChan := kubernetes.FilteredWatch(ctx, kubernetesProvider, targetNamespaces, mizu.Config.Tap.PodRegex())

	restartTappers := func() {
		err, changeFound := updateCurrentlyTappedPods(kubernetesProvider, ctx, targetNamespaces)
		if err != nil {
			mizu.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error getting pods by regex: %v", errormessage.FormatError(err)))
			cancel()
		}

		if !changeFound {
			mizu.Log.Debugf("Nothing changed update tappers not needed")
			return
		}

		reportTappedPods()

		nodeToTappedPodIPMap := getNodeHostToTappedPodIpsMap(state.currentlyTappedPods)
		if err != nil {
			mizu.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error building node to ips map: %v", errormessage.FormatError(err)))
			cancel()
		}
		if err := updateMizuTappers(ctx, kubernetesProvider, nodeToTappedPodIPMap); err != nil {
			mizu.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error updating daemonset: %v", errormessage.FormatError(err)))
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

func updateCurrentlyTappedPods(kubernetesProvider *kubernetes.Provider, ctx context.Context, targetNamespaces []string) (error, bool) {
	changeFound := false
	if matchingPods, err := kubernetesProvider.ListAllRunningPodsMatchingRegex(ctx, mizu.Config.Tap.PodRegex(), targetNamespaces); err != nil {
		return err, false
	} else {
		addedPods, removedPods := getPodArrayDiff(state.currentlyTappedPods, matchingPods)
		for _, addedPod := range addedPods {
			changeFound = true
			mizu.Log.Infof(uiUtils.Green, fmt.Sprintf("+%s", addedPod.Name))
		}
		for _, removedPod := range removedPods {
			changeFound = true
			mizu.Log.Infof(uiUtils.Red, fmt.Sprintf("-%s", removedPod.Name))
		}
		state.currentlyTappedPods = matchingPods
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
	added, modified, removed, errorChan := kubernetes.FilteredWatch(ctx, kubernetesProvider, []string{mizu.Config.MizuResourcesNamespace}, podExactRegex)
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
					err := kubernetes.StartProxy(kubernetesProvider, mizu.Config.Tap.GuiPort, mizu.Config.MizuResourcesNamespace, mizu.ApiServerPodName)
					if err != nil {
						mizu.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error occured while running k8s proxy %v", errormessage.FormatError(err)))
						cancel()
					}
				}()
				mizuProxiedUrl := kubernetes.GetMizuApiServerProxiedHostAndPath(mizu.Config.Tap.GuiPort)
				mizu.Log.Infof("Mizu is available at http://%s\n", mizuProxiedUrl)

				time.Sleep(time.Second * 5) // Waiting to be sure the proxy is ready
				requestForAnalysis()
				reportTappedPods()
			}
		case <-timeAfter:
			if !isPodReady {
				mizu.Log.Errorf(uiUtils.Error, fmt.Sprintf("%s pod was not ready in time", mizu.ApiServerPodName))
				cancel()
			}
		case <-errorChan:
			mizu.Log.Debugf("[ERROR] Agent creation, watching %v namespace", mizu.Config.MizuResourcesNamespace)
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

func createRBACIfNecessary(ctx context.Context, kubernetesProvider *kubernetes.Provider) (bool, error) {
	mizuRBACExists, err := kubernetesProvider.DoesServiceAccountExist(ctx, mizu.Config.MizuResourcesNamespace, mizu.ServiceAccountName)
	if err != nil {
		return false, err
	}
	if !mizuRBACExists {
		if !mizu.Config.IsNsRestrictedMode() {
			err := kubernetesProvider.CreateMizuRBAC(ctx, mizu.Config.MizuResourcesNamespace, mizu.ServiceAccountName, mizu.ClusterRoleName, mizu.ClusterRoleBindingName, mizu.RBACVersion)
			if err != nil {
				return false, err
			}
		} else {
			err := kubernetesProvider.CreateMizuRBACNamespaceRestricted(ctx, mizu.Config.MizuResourcesNamespace, mizu.ServiceAccountName, mizu.RoleName, mizu.RoleBindingName, mizu.RBACVersion)
			if err != nil {
				return false, err
			}
		}
	}
	return true, nil
}

func getNodeHostToTappedPodIpsMap(tappedPods []core.Pod) map[string][]string {
	nodeToTappedPodIPMap := make(map[string][]string, 0)
	for _, pod := range tappedPods {
		existingList := nodeToTappedPodIPMap[pod.Spec.NodeName]
		if existingList == nil {
			nodeToTappedPodIPMap[pod.Spec.NodeName] = []string{pod.Status.PodIP}
		} else {
			nodeToTappedPodIPMap[pod.Spec.NodeName] = append(nodeToTappedPodIPMap[pod.Spec.NodeName], pod.Status.PodIP)
		}
	}
	return nodeToTappedPodIPMap
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

func getNamespaces(kubernetesProvider *kubernetes.Provider) []string {
	if mizu.Config.Tap.AllNamespaces {
		return []string{mizu.K8sAllNamespaces}
	} else if len(mizu.Config.Tap.Namespaces) > 0 {
		return mizu.Config.Tap.Namespaces
	} else {
		return []string{kubernetesProvider.CurrentNamespace()}
	}
}
