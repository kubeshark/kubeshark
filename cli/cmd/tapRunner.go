package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/logger"
	"github.com/up9inc/mizu/cli/mizu/fsUtils"
	"github.com/up9inc/mizu/cli/mizu/goUtils"
	"github.com/up9inc/mizu/cli/mizu/version"
	"github.com/up9inc/mizu/cli/telemetry"
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
		logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error parsing regex-masking: %v", errormessage.FormatError(err)))
		return
	}
	var mizuValidationRules string
	if config.Config.Tap.EnforcePolicyFile != "" {
		mizuValidationRules, err = readValidationRules(config.Config.Tap.EnforcePolicyFile)
		if err != nil {
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error reading policy file: %v", errormessage.FormatError(err)))
			return
		}
	}

	kubernetesProvider, err := kubernetes.NewProvider(config.Config.KubeConfigPath)
	if err != nil {
		logger.Log.Error(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	targetNamespaces := getNamespaces(kubernetesProvider)

	var namespacesStr string
	if !mizu.Contains(targetNamespaces, mizu.K8sAllNamespaces) {
		namespacesStr = fmt.Sprintf("namespaces \"%s\"", strings.Join(targetNamespaces, "\", \""))
	} else {
		namespacesStr = "all namespaces"
	}
	version.CheckNewerVersion()
	logger.Log.Infof("Tapping pods in %s", namespacesStr)

	if err, _ := updateCurrentlyTappedPods(kubernetesProvider, ctx, targetNamespaces); err != nil {
		logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error getting pods by regex: %v", errormessage.FormatError(err)))
		return
	}

	if len(state.currentlyTappedPods) == 0 {
		var suggestionStr string
		if !mizu.Contains(targetNamespaces, mizu.K8sAllNamespaces) {
			suggestionStr = ". Select a different namespace with -n or tap all namespaces with -A"
		}
		logger.Log.Warningf(uiUtils.Warning, fmt.Sprintf("Did not find any pods matching the regex argument%s", suggestionStr))
	}

	if config.Config.Tap.DryRun {
		return
	}

	nodeToTappedPodIPMap := getNodeHostToTappedPodIpsMap(state.currentlyTappedPods)

	defer cleanUpMizu(kubernetesProvider)
	if err := createMizuResources(ctx, kubernetesProvider, nodeToTappedPodIPMap, mizuApiFilteringOptions, mizuValidationRules); err != nil {
		logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error creating resources: %v", errormessage.FormatError(err)))
		return
	}

	go goUtils.HandleExcWrapper(createProxyToApiServerPod, ctx, kubernetesProvider, cancel)
	go goUtils.HandleExcWrapper(watchPodsForTapping, ctx, kubernetesProvider, targetNamespaces, cancel)

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
	if !config.Config.IsNsRestrictedMode() {
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
		logger.Log.Warningf(uiUtils.Warning, fmt.Sprintf("Failed to create resources required for policy validation. Mizu will not validate policy rules. error: %v\n", errormessage.FormatError(err)))
		state.doNotRemoveConfigMap = true
	} else if mizuValidationRules == "" {
		state.doNotRemoveConfigMap = true
	}

	return nil
}

func createMizuConfigmap(ctx context.Context, kubernetesProvider *kubernetes.Provider, data string) error {
	err := kubernetesProvider.CreateConfigMap(ctx, config.Config.MizuResourcesNamespace, mizu.ConfigMapName, data)
	return err
}

func createMizuNamespace(ctx context.Context, kubernetesProvider *kubernetes.Provider) error {
	_, err := kubernetesProvider.CreateNamespace(ctx, config.Config.MizuResourcesNamespace)
	return err
}

func createMizuApiServer(ctx context.Context, kubernetesProvider *kubernetes.Provider, mizuApiFilteringOptions *shared.TrafficFilteringOptions) error {
	var err error

	state.mizuServiceAccountExists, err = createRBACIfNecessary(ctx, kubernetesProvider)
	if err != nil {
		logger.Log.Warningf(uiUtils.Warning, fmt.Sprintf("Failed to ensure the resources required for IP resolving. Mizu will not resolve target IPs to names. error: %v", errormessage.FormatError(err)))
	}

	var serviceAccountName string
	if state.mizuServiceAccountExists {
		serviceAccountName = mizu.ServiceAccountName
	} else {
		serviceAccountName = ""
	}

	opts := &kubernetes.ApiServerOptions{
		Namespace:               config.Config.MizuResourcesNamespace,
		PodName:                 mizu.ApiServerPodName,
		PodImage:                config.Config.AgentImage,
		ServiceAccountName:      serviceAccountName,
		IsNamespaceRestricted:   config.Config.IsNsRestrictedMode(),
		MizuApiFilteringOptions: mizuApiFilteringOptions,
		MaxEntriesDBSizeBytes:   config.Config.Tap.MaxEntriesDBSizeBytes(),
	}
	_, err = kubernetesProvider.CreateMizuApiServerPod(ctx, opts)
	if err != nil {
		return err
	}
	logger.Log.Debugf("Successfully created API server pod: %s", mizu.ApiServerPodName)

	state.apiServerService, err = kubernetesProvider.CreateService(ctx, config.Config.MizuResourcesNamespace, mizu.ApiServerPodName, mizu.ApiServerPodName)
	if err != nil {
		return err
	}
	logger.Log.Debugf("Successfully created service: %s", mizu.ApiServerPodName)

	return nil
}

func getMizuApiFilteringOptions() (*shared.TrafficFilteringOptions, error) {
	var compiledRegexSlice []*shared.SerializableRegexp

	if config.Config.Tap.PlainTextFilterRegexes != nil && len(config.Config.Tap.PlainTextFilterRegexes) > 0 {
		compiledRegexSlice = make([]*shared.SerializableRegexp, 0)
		for _, regexStr := range config.Config.Tap.PlainTextFilterRegexes {
			compiledRegex, err := shared.CompileRegexToSerializableRegexp(regexStr)
			if err != nil {
				return nil, err
			}
			compiledRegexSlice = append(compiledRegexSlice, compiledRegex)
		}
	}

	return &shared.TrafficFilteringOptions{
		PlainTextMaskingRegexes:      compiledRegexSlice,
		HealthChecksUserAgentHeaders: config.Config.Tap.HealthChecksUserAgentHeaders,
		DisableRedaction:             config.Config.Tap.DisableRedaction,
	}, nil
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
			config.Config.MizuResourcesNamespace,
			mizu.TapperDaemonSetName,
			config.Config.AgentImage,
			mizu.TapperPodName,
			fmt.Sprintf("%s.%s.svc.cluster.local", state.apiServerService.Name, state.apiServerService.Namespace),
			nodeToTappedPodIPMap,
			serviceAccountName,
			config.Config.Tap.TapOutgoing(),
		); err != nil {
			return err
		}
		logger.Log.Debugf("Successfully created %v tappers", len(nodeToTappedPodIPMap))
	} else {
		if err := kubernetesProvider.RemoveDaemonSet(ctx, config.Config.MizuResourcesNamespace, mizu.TapperDaemonSetName); err != nil {
			return err
		}
	}

	return nil
}

func cleanUpMizu(kubernetesProvider *kubernetes.Provider) {
	telemetry.ReportAPICalls(config.Config.Tap.GuiPort)
	cleanUpMizuResources(kubernetesProvider)
}

func cleanUpMizuResources(kubernetesProvider *kubernetes.Provider) {
	removalCtx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()

	if config.Config.DumpLogs {
		mizuDir := mizu.GetMizuFolderPath()
		filePath = path.Join(mizuDir, fmt.Sprintf("mizu_logs_%s.zip", time.Now().Format("2006_01_02__15_04_05")))
		if err := fsUtils.DumpLogs(kubernetesProvider, removalCtx, filePath); err != nil {
			logger.Log.Errorf("Failed dump logs %v", err)
		}
	}

	logger.Log.Infof("\nRemoving mizu resources\n")

	if !config.Config.IsNsRestrictedMode() {
		if err := kubernetesProvider.RemoveNamespace(removalCtx, config.Config.MizuResourcesNamespace); err != nil {
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error removing Namespace %s: %v", config.Config.MizuResourcesNamespace, errormessage.FormatError(err)))
			return
		}
	} else {
		if err := kubernetesProvider.RemovePod(removalCtx, config.Config.MizuResourcesNamespace, mizu.ApiServerPodName); err != nil {
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error removing Pod %s in namespace %s: %v", mizu.ApiServerPodName, config.Config.MizuResourcesNamespace, errormessage.FormatError(err)))
		}

		if err := kubernetesProvider.RemoveService(removalCtx, config.Config.MizuResourcesNamespace, mizu.ApiServerPodName); err != nil {
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error removing Service %s in namespace %s: %v", mizu.ApiServerPodName, config.Config.MizuResourcesNamespace, errormessage.FormatError(err)))
		}

		if err := kubernetesProvider.RemoveDaemonSet(removalCtx, config.Config.MizuResourcesNamespace, mizu.TapperDaemonSetName); err != nil {
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error removing DaemonSet %s in namespace %s: %v", mizu.TapperDaemonSetName, config.Config.MizuResourcesNamespace, errormessage.FormatError(err)))
		}

		if !state.doNotRemoveConfigMap {
			if err := kubernetesProvider.RemoveConfigMap(removalCtx, config.Config.MizuResourcesNamespace, mizu.ConfigMapName); err != nil {
				logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error removing ConfigMap %s in namespace %s: %v", mizu.ConfigMapName, config.Config.MizuResourcesNamespace, errormessage.FormatError(err)))
			}
		}

	}

	if state.mizuServiceAccountExists {
		if !config.Config.IsNsRestrictedMode() {
			if err := kubernetesProvider.RemoveNonNamespacedResources(removalCtx, mizu.ClusterRoleName, mizu.ClusterRoleBindingName); err != nil {
				logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error removing non-namespaced resources: %v", errormessage.FormatError(err)))
				return
			}
		} else {
			if err := kubernetesProvider.RemoveServicAccount(removalCtx, config.Config.MizuResourcesNamespace, mizu.ServiceAccountName); err != nil {
				logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error removing Service Account %s in namespace %s: %v", mizu.ServiceAccountName, config.Config.MizuResourcesNamespace, errormessage.FormatError(err)))
				return
			}

			if err := kubernetesProvider.RemoveRole(removalCtx, config.Config.MizuResourcesNamespace, mizu.RoleName); err != nil {
				logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error removing Role %s in namespace %s: %v", mizu.RoleName, config.Config.MizuResourcesNamespace, errormessage.FormatError(err)))
			}

			if err := kubernetesProvider.RemoveRoleBinding(removalCtx, config.Config.MizuResourcesNamespace, mizu.RoleBindingName); err != nil {
				logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error removing RoleBinding %s in namespace %s: %v", mizu.RoleBindingName, config.Config.MizuResourcesNamespace, errormessage.FormatError(err)))
			}
		}
	}

	if !config.Config.IsNsRestrictedMode() {
		waitUntilNamespaceDeleted(removalCtx, cancel, kubernetesProvider)
	}
}

func waitUntilNamespaceDeleted(ctx context.Context, cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider) {
	// Call cancel if a terminating signal was received. Allows user to skip the wait.
	go func() {
		waitForFinish(ctx, cancel)
	}()

	if err := kubernetesProvider.WaitUtilNamespaceDeleted(ctx, config.Config.MizuResourcesNamespace); err != nil {
		switch {
		case ctx.Err() == context.Canceled:
			// Do nothing. User interrupted the wait.
		case err == wait.ErrWaitTimeout:
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Timeout while removing Namespace %s", config.Config.MizuResourcesNamespace))
		default:
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error while waiting for Namespace %s to be deleted: %v", config.Config.MizuResourcesNamespace, errormessage.FormatError(err)))
		}
	}
}

func reportTappedPods() {
	mizuProxiedUrl := kubernetes.GetMizuApiServerProxiedHostAndPath(config.Config.Tap.GuiPort)
	tappedPodsUrl := fmt.Sprintf("http://%s/status/tappedPods", mizuProxiedUrl)

	podInfos := make([]shared.PodInfo, 0)
	for _, pod := range state.currentlyTappedPods {
		podInfos = append(podInfos, shared.PodInfo{Name: pod.Name, Namespace: pod.Namespace})
	}
	tapStatus := shared.TapStatus{Pods: podInfos}

	if jsonValue, err := json.Marshal(tapStatus); err != nil {
		logger.Log.Debugf("[ERROR] failed Marshal the tapped pods %v", err)
	} else {
		if response, err := http.Post(tappedPodsUrl, "application/json", bytes.NewBuffer(jsonValue)); err != nil {
			logger.Log.Debugf("[ERROR] failed sending to API server the tapped pods %v", err)
		} else if response.StatusCode != 200 {
			logger.Log.Debugf("[ERROR] failed sending to API server the tapped pods, response status code %v", response.StatusCode)
		} else {
			logger.Log.Debugf("Reported to server API about %d taped pods successfully", len(podInfos))
		}
	}
}

func watchPodsForTapping(ctx context.Context, kubernetesProvider *kubernetes.Provider, targetNamespaces []string, cancel context.CancelFunc) {
	added, modified, removed, errorChan := kubernetes.FilteredWatch(ctx, kubernetesProvider, targetNamespaces, config.Config.Tap.PodRegex())

	restartTappers := func() {
		err, changeFound := updateCurrentlyTappedPods(kubernetesProvider, ctx, targetNamespaces)
		if err != nil {
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Failed to update currently tapped pods: %v", err))
			cancel()
		}

		if !changeFound {
			logger.Log.Debugf("Nothing changed update tappers not needed")
			return
		}

		reportTappedPods()

		nodeToTappedPodIPMap := getNodeHostToTappedPodIpsMap(state.currentlyTappedPods)
		if err != nil {
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error building node to ips map: %v", errormessage.FormatError(err)))
			cancel()
		}
		if err := updateMizuTappers(ctx, kubernetesProvider, nodeToTappedPodIPMap); err != nil {
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error updating daemonset: %v", errormessage.FormatError(err)))
			cancel()
		}
	}
	restartTappersDebouncer := debounce.NewDebouncer(updateTappersDelay, restartTappers)

	for {
		select {
		case pod := <-added:
			logger.Log.Debugf("Added matching pod %s, ns: %s", pod.Name, pod.Namespace)
			restartTappersDebouncer.SetOn()
		case pod := <-removed:
			logger.Log.Debugf("Removed matching pod %s, ns: %s", pod.Name, pod.Namespace)
			restartTappersDebouncer.SetOn()
		case pod := <-modified:
			logger.Log.Debugf("Modified matching pod %s, ns: %s, phase: %s, ip: %s", pod.Name, pod.Namespace, pod.Status.Phase, pod.Status.PodIP)
			// Act only if the modified pod has already obtained an IP address.
			// After filtering for IPs, on a normal pod restart this includes the following events:
			// - Pod deletion
			// - Pod reaches start state
			// - Pod reaches ready state
			// Ready/unready transitions might also trigger this event.
			if pod.Status.PodIP != "" {
				restartTappersDebouncer.SetOn()
			}

		case err := <-errorChan:
			logger.Log.Debugf("Watching pods loop, got error %v, stopping `restart tappers debouncer`", err)
			restartTappersDebouncer.Cancel()
			// TODO: Does this also perform cleanup?
			cancel()

		case <-ctx.Done():
			logger.Log.Debugf("Watching pods loop, context done, stopping `restart tappers debouncer`")
			restartTappersDebouncer.Cancel()
			return
		}
	}
}

func updateCurrentlyTappedPods(kubernetesProvider *kubernetes.Provider, ctx context.Context, targetNamespaces []string) (error, bool) {
	changeFound := false
	if matchingPods, err := kubernetesProvider.ListAllRunningPodsMatchingRegex(ctx, config.Config.Tap.PodRegex(), targetNamespaces); err != nil {
		return err, false
	} else {
		podsToTap := excludeMizuPods(matchingPods)
		addedPods, removedPods := getPodArrayDiff(state.currentlyTappedPods, podsToTap)
		for _, addedPod := range addedPods {
			changeFound = true
			logger.Log.Infof(uiUtils.Green, fmt.Sprintf("+%s", addedPod.Name))
		}
		for _, removedPod := range removedPods {
			changeFound = true
			logger.Log.Infof(uiUtils.Red, fmt.Sprintf("-%s", removedPod.Name))
		}
		state.currentlyTappedPods = podsToTap
	}

	return nil, changeFound
}

func excludeMizuPods(pods []core.Pod) []core.Pod {
	mizuPrefixRegex := regexp.MustCompile("^" + mizu.MizuResourcesPrefix)

	nonMizuPods := make([]core.Pod, 0)
	for _, pod := range pods {
		if !mizuPrefixRegex.MatchString(pod.Name) {
			nonMizuPods = append(nonMizuPods, pod)
		}
	}

	return nonMizuPods
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
	added, modified, removed, errorChan := kubernetes.FilteredWatch(ctx, kubernetesProvider, []string{config.Config.MizuResourcesNamespace}, podExactRegex)
	isPodReady := false
	timeAfter := time.After(25 * time.Second)
	for {
		select {
		case <-ctx.Done():
			logger.Log.Debugf("Watching API Server pod loop, ctx done")
			return
		case <-added:
			logger.Log.Debugf("Watching API Server pod loop, added")
			continue
		case <-removed:
			logger.Log.Infof("%s removed", mizu.ApiServerPodName)
			cancel()
			return
		case modifiedPod := <-modified:
			if modifiedPod == nil {
				logger.Log.Debugf("Watching API Server pod loop, modifiedPod with nil")
				continue
			}
			logger.Log.Debugf("Watching API Server pod loop, modified: %v", modifiedPod.Status.Phase)
			if modifiedPod.Status.Phase == core.PodRunning && !isPodReady {
				isPodReady = true
				go startProxyReportErrorIfAny(kubernetesProvider, cancel)
				logger.Log.Infof("Mizu is available at http://%s\n", kubernetes.GetMizuApiServerProxiedHostAndPath(config.Config.Tap.GuiPort))
				time.Sleep(time.Second * 5) // Waiting to be sure the proxy is ready
				requestForAnalysis()
				reportTappedPods()
			}
		case <-timeAfter:
			if !isPodReady {
				logger.Log.Errorf(uiUtils.Error, "Mizu API server was not ready in time")
				cancel()
			}
		case <-errorChan:
			logger.Log.Debugf("[ERROR] Agent creation, watching %v namespace", config.Config.MizuResourcesNamespace)
			cancel()
		}
	}
}

func startProxyReportErrorIfAny(kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc) {
	err := kubernetes.StartProxy(kubernetesProvider, config.Config.Tap.GuiPort, config.Config.MizuResourcesNamespace, mizu.ApiServerPodName)
	if err != nil {
		logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error occured while running k8s proxy %v\n"+
			"Try setting different port by using --%s", errormessage.FormatError(err), configStructs.GuiPortTapName))
		cancel()
	}
}

func requestForAnalysis() {
	if !config.Config.Tap.Analysis {
		return
	}

	mizuProxiedUrl := kubernetes.GetMizuApiServerProxiedHostAndPath(config.Config.Tap.GuiPort)
	urlPath := fmt.Sprintf("http://%s/api/uploadEntries?dest=%s&interval=%v", mizuProxiedUrl, url.QueryEscape(config.Config.Tap.AnalysisDestination), config.Config.Tap.SleepIntervalSec)
	u, parseErr := url.ParseRequestURI(urlPath)
	if parseErr != nil {
		logger.Log.Fatal("Failed parsing the URL (consider changing the analysis dest URL), err: %v", parseErr)
	}

	logger.Log.Debugf("Sending get request to %v", u.String())
	if response, requestErr := http.Get(u.String()); requestErr != nil {
		logger.Log.Errorf("Failed to notify agent for analysis, err: %v", requestErr)
	} else if response.StatusCode != 200 {
		logger.Log.Errorf("Failed to notify agent for analysis, status code: %v", response.StatusCode)
	} else {
		logger.Log.Infof(uiUtils.Purple, "Traffic is uploading to UP9 for further analysis")
	}
}

func createRBACIfNecessary(ctx context.Context, kubernetesProvider *kubernetes.Provider) (bool, error) {
	if !config.Config.IsNsRestrictedMode() {
		err := kubernetesProvider.CreateMizuRBAC(ctx, config.Config.MizuResourcesNamespace, mizu.ServiceAccountName, mizu.ClusterRoleName, mizu.ClusterRoleBindingName, mizu.RBACVersion)
		if err != nil {
			return false, err
		}
	} else {
		err := kubernetesProvider.CreateMizuRBACNamespaceRestricted(ctx, config.Config.MizuResourcesNamespace, mizu.ServiceAccountName, mizu.RoleName, mizu.RoleBindingName, mizu.RBACVersion)
		if err != nil {
			return false, err
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
	if config.Config.Tap.AllNamespaces {
		return []string{mizu.K8sAllNamespaces}
	} else if len(config.Config.Tap.Namespaces) > 0 {
		return config.Config.Tap.Namespaces
	} else {
		return []string{kubernetesProvider.CurrentNamespace()}
	}
}
