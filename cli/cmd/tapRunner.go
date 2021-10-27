package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/up9inc/mizu/cli/apiserver"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/errormessage"

	"github.com/up9inc/mizu/cli/kubernetes"
	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/cli/mizu/fsUtils"
	"github.com/up9inc/mizu/cli/mizu/goUtils"
	"github.com/up9inc/mizu/cli/telemetry"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/debounce"
	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/tap/api"
)

const (
	cleanupTimeout     = time.Minute
	updateTappersDelay = 5 * time.Second
)

type tapState struct {
	apiServerService         *core.Service
	currentlyTappedPods      []core.Pod
	mizuServiceAccountExists bool
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

	// Read and validate the OAS file
	var contract string
	if config.Config.Tap.ContractFile != "" {
		bytes, err := ioutil.ReadFile(config.Config.Tap.ContractFile)
		if err != nil {
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error reading contract file: %v", errormessage.FormatError(err)))
			return
		}
		contract = string(bytes)

		ctx := context.Background()
		loader := &openapi3.Loader{Context: ctx}
		doc, err := loader.LoadFromData(bytes)
		if err != nil {
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error loading contract file: %v", errormessage.FormatError(err)))
			return
		}
		err = doc.Validate(ctx)
		if err != nil {
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error validating contract file: %v", errormessage.FormatError(err)))
			return
		}
	}

	kubernetesProvider, err := kubernetes.NewProvider(config.Config.KubeConfigPath())
	if err != nil {
		logger.Log.Error(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	targetNamespaces := getNamespaces(kubernetesProvider)

	if config.Config.IsNsRestrictedMode() {
		if len(targetNamespaces) != 1 || !shared.Contains(targetNamespaces, config.Config.MizuResourcesNamespace) {
			logger.Log.Errorf("Not supported mode. Mizu can't resolve IPs in other namespaces when running in namespace restricted mode.\n"+
				"You can use the same namespace for --%s and --%s", configStructs.NamespacesTapName, config.MizuResourcesNamespaceConfigName)
			return
		}
	}

	var namespacesStr string
	if !shared.Contains(targetNamespaces, mizu.K8sAllNamespaces) {
		namespacesStr = fmt.Sprintf("namespaces \"%s\"", strings.Join(targetNamespaces, "\", \""))
	} else {
		namespacesStr = "all namespaces"
	}

	logger.Log.Infof("Tapping pods in %s", namespacesStr)

	if err, _ := updateCurrentlyTappedPods(kubernetesProvider, ctx, targetNamespaces); err != nil {
		logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error getting pods by regex: %v", errormessage.FormatError(err)))
		return
	}

	if len(state.currentlyTappedPods) == 0 {
		var suggestionStr string
		if !shared.Contains(targetNamespaces, mizu.K8sAllNamespaces) {
			suggestionStr = ". Select a different namespace with -n or tap all namespaces with -A"
		}
		logger.Log.Warningf(uiUtils.Warning, fmt.Sprintf("Did not find any pods matching the regex argument%s", suggestionStr))
	}

	if config.Config.Tap.DryRun {
		return
	}

	if err := createMizuResources(ctx, kubernetesProvider, mizuValidationRules, contract); err != nil {
		logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error creating resources: %v", errormessage.FormatError(err)))

		var statusError *k8serrors.StatusError
		if errors.As(err, &statusError) {
			if statusError.ErrStatus.Reason == metav1.StatusReasonAlreadyExists {
				logger.Log.Info("Mizu is already running in this namespace, change the `mizu-resources-namespace` configuration or run `mizu clean` to remove the currently running Mizu instance")
			}
		}
		return
	}
	defer finishMizuExecution(kubernetesProvider)

	go goUtils.HandleExcWrapper(watchApiServerPod, ctx, kubernetesProvider, cancel, mizuApiFilteringOptions)
	go goUtils.HandleExcWrapper(watchTapperPod, ctx, kubernetesProvider, cancel)
	go goUtils.HandleExcWrapper(watchPodsForTapping, ctx, kubernetesProvider, targetNamespaces, cancel, mizuApiFilteringOptions)

	// block until exit signal or error
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

func createMizuResources(ctx context.Context, kubernetesProvider *kubernetes.Provider, mizuValidationRules string, contract string) error {
	if !config.Config.IsNsRestrictedMode() {
		if err := createMizuNamespace(ctx, kubernetesProvider); err != nil {
			return err
		}
	}

	if err := createMizuApiServer(ctx, kubernetesProvider); err != nil {
		return err
	}

	mizuConfig, err := getSerializedMizuConfig()
	if err != nil {
		return err
	}

	if err := createMizuConfigmap(ctx, kubernetesProvider, mizuValidationRules, contract, mizuConfig); err != nil {
		logger.Log.Warningf(uiUtils.Warning, fmt.Sprintf("Failed to create resources required for policy validation. Mizu will not validate policy rules. error: %v\n", errormessage.FormatError(err)))
	}

	return nil
}

func createMizuConfigmap(ctx context.Context, kubernetesProvider *kubernetes.Provider, data string, contract string, mizuConfig string) error {
	err := kubernetesProvider.CreateConfigMap(ctx, config.Config.MizuResourcesNamespace, mizu.ConfigMapName, data, contract, mizuConfig)
	return err
}

func createMizuNamespace(ctx context.Context, kubernetesProvider *kubernetes.Provider) error {
	_, err := kubernetesProvider.CreateNamespace(ctx, config.Config.MizuResourcesNamespace)
	return err
}

func createMizuApiServer(ctx context.Context, kubernetesProvider *kubernetes.Provider) error {
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
		Namespace:             config.Config.MizuResourcesNamespace,
		PodName:               mizu.ApiServerPodName,
		PodImage:              config.Config.AgentImage,
		ServiceAccountName:    serviceAccountName,
		IsNamespaceRestricted: config.Config.IsNsRestrictedMode(),
		SyncEntriesConfig:     getSyncEntriesConfig(),
		MaxEntriesDBSizeBytes: config.Config.Tap.MaxEntriesDBSizeBytes(),
		Resources:             config.Config.Tap.ApiServerResources,
		ImagePullPolicy:       config.Config.ImagePullPolicy(),
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

func getMizuApiFilteringOptions() (*api.TrafficFilteringOptions, error) {
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

	return &api.TrafficFilteringOptions{
		PlainTextMaskingRegexes: compiledRegexSlice,
		IgnoredUserAgents:       config.Config.Tap.IgnoredUserAgents,
		DisableRedaction:        config.Config.Tap.DisableRedaction,
	}, nil
}

func getSyncEntriesConfig() *shared.SyncEntriesConfig {
	if !config.Config.Tap.Analysis && config.Config.Tap.Workspace == "" {
		return nil
	}

	return &shared.SyncEntriesConfig{
		Token:             config.Config.Auth.Token,
		Env:               config.Config.Auth.EnvName,
		Workspace:         config.Config.Tap.Workspace,
		UploadIntervalSec: config.Config.Tap.UploadIntervalSec,
	}
}

func updateMizuTappers(ctx context.Context, kubernetesProvider *kubernetes.Provider, mizuApiFilteringOptions *api.TrafficFilteringOptions) error {
	nodeToTappedPodIPMap := getNodeHostToTappedPodIpsMap(state.currentlyTappedPods)

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
			config.Config.Tap.TapperResources,
			config.Config.ImagePullPolicy(),
			mizuApiFilteringOptions,
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

func finishMizuExecution(kubernetesProvider *kubernetes.Provider) {
	telemetry.ReportAPICalls()
	removalCtx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()
	dumpLogsIfNeeded(removalCtx, kubernetesProvider)
	cleanUpMizuResources(removalCtx, cancel, kubernetesProvider)
}

func dumpLogsIfNeeded(ctx context.Context, kubernetesProvider *kubernetes.Provider) {
	if !config.Config.DumpLogs {
		return
	}
	mizuDir := mizu.GetMizuFolderPath()
	filePath := path.Join(mizuDir, fmt.Sprintf("mizu_logs_%s.zip", time.Now().Format("2006_01_02__15_04_05")))
	if err := fsUtils.DumpLogs(ctx, kubernetesProvider, filePath); err != nil {
		logger.Log.Errorf("Failed dump logs %v", err)
	}
}

func cleanUpMizuResources(ctx context.Context, cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider) {
	logger.Log.Infof("\nRemoving mizu resources\n")

	var leftoverResources []string

	if config.Config.IsNsRestrictedMode() {
		leftoverResources = cleanUpRestrictedMode(ctx, kubernetesProvider)
	} else {
		leftoverResources = cleanUpNonRestrictedMode(ctx, cancel, kubernetesProvider)
	}

	if len(leftoverResources) > 0 {
		errMsg := fmt.Sprintf("Failed to remove the following resources, for more info check logs at %s:", fsUtils.GetLogFilePath())
		for _, resource := range leftoverResources {
			errMsg += "\n- " + resource
		}
		logger.Log.Errorf(uiUtils.Error, errMsg)
	}
}

func cleanUpRestrictedMode(ctx context.Context, kubernetesProvider *kubernetes.Provider) []string {
	leftoverResources := make([]string, 0)

	if err := kubernetesProvider.RemovePod(ctx, config.Config.MizuResourcesNamespace, mizu.ApiServerPodName); err != nil {
		resourceDesc := fmt.Sprintf("Pod %s in namespace %s", mizu.ApiServerPodName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveService(ctx, config.Config.MizuResourcesNamespace, mizu.ApiServerPodName); err != nil {
		resourceDesc := fmt.Sprintf("Service %s in namespace %s", mizu.ApiServerPodName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveDaemonSet(ctx, config.Config.MizuResourcesNamespace, mizu.TapperDaemonSetName); err != nil {
		resourceDesc := fmt.Sprintf("DaemonSet %s in namespace %s", mizu.TapperDaemonSetName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveConfigMap(ctx, config.Config.MizuResourcesNamespace, mizu.ConfigMapName); err != nil {
		resourceDesc := fmt.Sprintf("ConfigMap %s in namespace %s", mizu.ConfigMapName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveServicAccount(ctx, config.Config.MizuResourcesNamespace, mizu.ServiceAccountName); err != nil {
		resourceDesc := fmt.Sprintf("Service Account %s in namespace %s", mizu.ServiceAccountName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveRole(ctx, config.Config.MizuResourcesNamespace, mizu.RoleName); err != nil {
		resourceDesc := fmt.Sprintf("Role %s in namespace %s", mizu.RoleName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveRoleBinding(ctx, config.Config.MizuResourcesNamespace, mizu.RoleBindingName); err != nil {
		resourceDesc := fmt.Sprintf("RoleBinding %s in namespace %s", mizu.RoleBindingName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	return leftoverResources
}

func cleanUpNonRestrictedMode(ctx context.Context, cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider) []string {
	leftoverResources := make([]string, 0)

	if err := kubernetesProvider.RemoveNamespace(ctx, config.Config.MizuResourcesNamespace); err != nil {
		resourceDesc := fmt.Sprintf("Namespace %s", config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	} else {
		defer waitUntilNamespaceDeleted(ctx, cancel, kubernetesProvider)
	}

	if err := kubernetesProvider.RemoveClusterRole(ctx, mizu.ClusterRoleName); err != nil {
		resourceDesc := fmt.Sprintf("ClusterRole %s", mizu.ClusterRoleName)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveClusterRoleBinding(ctx, mizu.ClusterRoleBindingName); err != nil {
		resourceDesc := fmt.Sprintf("ClusterRoleBinding %s", mizu.ClusterRoleBindingName)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	return leftoverResources
}

func handleDeletionError(err error, resourceDesc string, leftoverResources *[]string) {
	logger.Log.Debugf("Error removing %s: %v", resourceDesc, errormessage.FormatError(err))
	*leftoverResources = append(*leftoverResources, resourceDesc)
}

func waitUntilNamespaceDeleted(ctx context.Context, cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider) {
	// Call cancel if a terminating signal was received. Allows user to skip the wait.
	go func() {
		waitForFinish(ctx, cancel)
	}()

	if err := kubernetesProvider.WaitUtilNamespaceDeleted(ctx, config.Config.MizuResourcesNamespace); err != nil {
		switch {
		case ctx.Err() == context.Canceled:
			logger.Log.Debugf("Do nothing. User interrupted the wait")
		case err == wait.ErrWaitTimeout:
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Timeout while removing Namespace %s", config.Config.MizuResourcesNamespace))
		default:
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error while waiting for Namespace %s to be deleted: %v", config.Config.MizuResourcesNamespace, errormessage.FormatError(err)))
		}
	}
}

func watchPodsForTapping(ctx context.Context, kubernetesProvider *kubernetes.Provider, targetNamespaces []string, cancel context.CancelFunc, mizuApiFilteringOptions *api.TrafficFilteringOptions) {
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

		if err := apiserver.Provider.ReportTappedPods(state.currentlyTappedPods); err != nil {
			logger.Log.Debugf("[Error] failed update tapped pods %v", err)
		}

		if err := updateMizuTappers(ctx, kubernetesProvider, mizuApiFilteringOptions); err != nil {
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error updating tappers: %v", errormessage.FormatError(err)))
			cancel()
		}
	}
	restartTappersDebouncer := debounce.NewDebouncer(updateTappersDelay, restartTappers)

	for {
		select {
		case pod, ok := <-added:
			if !ok {
				added = nil
				continue
			}

			logger.Log.Debugf("Added matching pod %s, ns: %s", pod.Name, pod.Namespace)
			restartTappersDebouncer.SetOn()
		case pod, ok := <-removed:
			if !ok {
				removed = nil
				continue
			}

			logger.Log.Debugf("Removed matching pod %s, ns: %s", pod.Name, pod.Namespace)
			restartTappersDebouncer.SetOn()
		case pod, ok := <-modified:
			if !ok {
				modified = nil
				continue
			}

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
		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
				continue
			}

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

func watchApiServerPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc, mizuApiFilteringOptions *api.TrafficFilteringOptions) {
	podExactRegex := regexp.MustCompile(fmt.Sprintf("^%s$", mizu.ApiServerPodName))
	added, modified, removed, errorChan := kubernetes.FilteredWatch(ctx, kubernetesProvider, []string{config.Config.MizuResourcesNamespace}, podExactRegex)
	isPodReady := false
	timeAfter := time.After(25 * time.Second)
	for {
		select {
		case _, ok := <-added:
			if !ok {
				added = nil
				continue
			}

			logger.Log.Debugf("Watching API Server pod loop, added")
		case _, ok := <-removed:
			if !ok {
				removed = nil
				continue
			}

			logger.Log.Infof("%s removed", mizu.ApiServerPodName)
			cancel()
			return
		case modifiedPod, ok := <-modified:
			if !ok {
				modified = nil
				continue
			}

			logger.Log.Debugf("Watching API Server pod loop, modified: %v", modifiedPod.Status.Phase)

			if modifiedPod.Status.Phase == core.PodPending {
				if modifiedPod.Status.Conditions[0].Type == core.PodScheduled && modifiedPod.Status.Conditions[0].Status != core.ConditionTrue {
					logger.Log.Debugf("Wasn't able to deploy the API server. Reason: \"%s\"", modifiedPod.Status.Conditions[0].Message)
					logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Wasn't able to deploy the API server, for more info check logs at %s", fsUtils.GetLogFilePath()))
					cancel()
					break
				}

				if len(modifiedPod.Status.ContainerStatuses) > 0 && modifiedPod.Status.ContainerStatuses[0].State.Waiting != nil && modifiedPod.Status.ContainerStatuses[0].State.Waiting.Reason == "ErrImagePull" {
					logger.Log.Debugf("Wasn't able to deploy the API server. (ErrImagePull) Reason: \"%s\"", modifiedPod.Status.ContainerStatuses[0].State.Waiting.Message)
					logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Wasn't able to deploy the API server: failed to pull the image, for more info check logs at %v", fsUtils.GetLogFilePath()))
					cancel()
					break
				}
			}

			if modifiedPod.Status.Phase == core.PodRunning && !isPodReady {
				isPodReady = true
				go startProxyReportErrorIfAny(kubernetesProvider, cancel)

				url := GetApiServerUrl()
				if err := apiserver.Provider.InitAndTestConnection(url); err != nil {
					logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Couldn't connect to API server, for more info check logs at %s", fsUtils.GetLogFilePath()))
					cancel()
					break
				}
				if err := updateMizuTappers(ctx, kubernetesProvider, mizuApiFilteringOptions); err != nil {
					logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error updating tappers: %v", errormessage.FormatError(err)))
					cancel()
				}

				logger.Log.Infof("Mizu is available at %s\n", url)
				uiUtils.OpenBrowser(url)
				if err := apiserver.Provider.ReportTappedPods(state.currentlyTappedPods); err != nil {
					logger.Log.Debugf("[Error] failed update tapped pods %v", err)
				}
			}
		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
				continue
			}

			logger.Log.Debugf("[ERROR] Agent creation, watching %v namespace, error: %v", config.Config.MizuResourcesNamespace, err)
			cancel()

		case <-timeAfter:
			if !isPodReady {
				logger.Log.Errorf(uiUtils.Error, "Mizu API server was not ready in time")
				cancel()
			}
		case <-ctx.Done():
			logger.Log.Debugf("Watching API Server pod loop, ctx done")
			return
		}
	}
}

func watchTapperPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc) {
	podExactRegex := regexp.MustCompile(fmt.Sprintf("^%s.*", mizu.TapperDaemonSetName))
	added, modified, removed, errorChan := kubernetes.FilteredWatch(ctx, kubernetesProvider, []string{config.Config.MizuResourcesNamespace}, podExactRegex)
	var prevPodPhase core.PodPhase
	for {
		select {
		case addedPod, ok := <-added:
			if !ok {
				added = nil
				continue
			}

			logger.Log.Debugf("Tapper is created [%s]", addedPod.Name)
		case removedPod, ok := <-removed:
			if !ok {
				removed = nil
				continue
			}

			logger.Log.Debugf("Tapper is removed [%s]", removedPod.Name)
		case modifiedPod, ok := <-modified:
			if !ok {
				modified = nil
				continue
			}

			if modifiedPod.Status.Phase == core.PodPending && modifiedPod.Status.Conditions[0].Type == core.PodScheduled && modifiedPod.Status.Conditions[0].Status != core.ConditionTrue {
				logger.Log.Infof(uiUtils.Red, fmt.Sprintf("Wasn't able to deploy the tapper %s. Reason: \"%s\"", modifiedPod.Name, modifiedPod.Status.Conditions[0].Message))
				cancel()
				break
			}

			podStatus := modifiedPod.Status
			if podStatus.Phase == core.PodPending && prevPodPhase == podStatus.Phase {
				logger.Log.Debugf("Tapper %s is %s", modifiedPod.Name, strings.ToLower(string(podStatus.Phase)))
				continue
			}
			prevPodPhase = podStatus.Phase

			if podStatus.Phase == core.PodRunning {
				state := podStatus.ContainerStatuses[0].State
				if state.Terminated != nil {
					switch state.Terminated.Reason {
					case "OOMKilled":
						logger.Log.Infof(uiUtils.Red, fmt.Sprintf("Tapper %s was terminated (reason: OOMKilled). You should consider increasing machine resources.", modifiedPod.Name))
					}
				}
			}

			logger.Log.Debugf("Tapper %s is %s", modifiedPod.Name, strings.ToLower(string(podStatus.Phase)))
		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
				continue
			}

			logger.Log.Debugf("[Error] Error in mizu tapper watch, err: %v", err)
			cancel()

		case <-ctx.Done():
			logger.Log.Debugf("Watching tapper pod loop, ctx done")
			return
		}
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

func getNamespaces(kubernetesProvider *kubernetes.Provider) []string {
	if config.Config.Tap.AllNamespaces {
		return []string{mizu.K8sAllNamespaces}
	} else if len(config.Config.Tap.Namespaces) > 0 {
		return shared.Unique(config.Config.Tap.Namespaces)
	} else {
		return []string{kubernetesProvider.CurrentNamespace()}
	}
}

func getSerializedMizuConfig() (string, error) {
	mizuConfig, err := getMizuConfig()
	if err != nil {
		return "", err
	}
	serializedConfig, err := json.Marshal(mizuConfig)
	if err != nil {
		return "", err
	}
	return string(serializedConfig), nil
}

func getMizuConfig() (*shared.MizuConfig, error) {
	serializableRegex, err := shared.CompileRegexToSerializableRegexp(config.Config.Tap.PodRegexStr)
	if err != nil {
		return nil, err
	}
	config := shared.MizuConfig{
		TapTargetRegex: *serializableRegex,
		MaxDBSizeBytes: config.Config.Tap.MaxEntriesDBSizeBytes(),
	}
	return &config, nil
}
