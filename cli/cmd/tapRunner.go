package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/up9inc/mizu/cli/cmd/goUtils"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/up9inc/mizu/cli/apiserver"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/errormessage"
	"gopkg.in/yaml.v3"
	core "k8s.io/api/core/v1"

	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/cli/mizu/fsUtils"
	"github.com/up9inc/mizu/cli/telemetry"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/kubernetes"
	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/tap/api"
)

const cleanupTimeout = time.Minute

type tapState struct {
	apiServerService         *core.Service
	tapperSyncer             *kubernetes.MizuTapperSyncer
	mizuServiceAccountExists bool
}

var state tapState
var apiProvider *apiserver.Provider

func RunMizuTap() {
	mizuApiFilteringOptions, err := getMizuApiFilteringOptions()
	apiProvider = apiserver.NewProvider(GetApiServerUrl(), apiserver.DefaultRetries, apiserver.DefaultTimeout)
	if err != nil {
		logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error parsing regex-masking: %v", errormessage.FormatError(err)))
		return
	}

	var serializedValidationRules string
	if config.Config.Tap.EnforcePolicyFile != "" {
		serializedValidationRules, err = readValidationRules(config.Config.Tap.EnforcePolicyFile)
		if err != nil {
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error reading policy file: %v", errormessage.FormatError(err)))
			return
		}
	}

	// Read and validate the OAS file
	var serializedContract string
	if config.Config.Tap.ContractFile != "" {
		bytes, err := ioutil.ReadFile(config.Config.Tap.ContractFile)
		if err != nil {
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error reading contract file: %v", errormessage.FormatError(err)))
			return
		}
		serializedContract = string(bytes)

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

	kubernetesProvider, err := getKubernetesProviderForCli()
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	targetNamespaces := getNamespaces(kubernetesProvider)

	serializedMizuConfig, err := config.GetSerializedMizuAgentConfig(targetNamespaces, mizuApiFilteringOptions)
	if err != nil {
		logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error composing mizu config: %v", errormessage.FormatError(err)))
		return
	}

	if config.Config.IsNsRestrictedMode() {
		if len(targetNamespaces) != 1 || !shared.Contains(targetNamespaces, config.Config.MizuResourcesNamespace) {
			logger.Log.Errorf("Not supported mode. Mizu can't resolve IPs in other namespaces when running in namespace restricted mode.\n"+
				"You can use the same namespace for --%s and --%s", configStructs.NamespacesTapName, config.MizuResourcesNamespaceConfigName)
			return
		}
	}

	var namespacesStr string
	if !shared.Contains(targetNamespaces, kubernetes.K8sAllNamespaces) {
		namespacesStr = fmt.Sprintf("namespaces \"%s\"", strings.Join(targetNamespaces, "\", \""))
	} else {
		namespacesStr = "all namespaces"
	}

	logger.Log.Infof("Tapping pods in %s", namespacesStr)

	if config.Config.Tap.DryRun {
		return
	}

	if err := createMizuResources(ctx, cancel, kubernetesProvider, serializedValidationRules, serializedContract, serializedMizuConfig); err != nil {
		logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error creating resources: %v", errormessage.FormatError(err)))

		var statusError *k8serrors.StatusError
		if errors.As(err, &statusError) {
			if statusError.ErrStatus.Reason == metav1.StatusReasonAlreadyExists {
				logger.Log.Info("Mizu is already running in this namespace, change the `mizu-resources-namespace` configuration or run `mizu clean` to remove the currently running Mizu instance")
			}
		}
		return
	}
	if config.Config.Tap.DaemonMode {
		if err := handleDaemonModePostCreation(cancel, kubernetesProvider); err != nil {
			defer finishMizuExecution(kubernetesProvider, apiProvider)
			cancel()
		} else {
			logger.Log.Infof(uiUtils.Magenta, "Mizu is now running in daemon mode, run `mizu view` to connect to the mizu daemon instance")
		}
	} else {
		defer finishMizuExecution(kubernetesProvider, apiProvider)

		if err = startTapperSyncer(ctx, cancel, kubernetesProvider, targetNamespaces, *mizuApiFilteringOptions); err != nil {
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error starting mizu tapper syncer: %v", err))
			cancel()
		}

		go goUtils.HandleExcWrapper(watchApiServerPod, ctx, kubernetesProvider, cancel)
		go goUtils.HandleExcWrapper(watchTapperPod, ctx, kubernetesProvider, cancel)

		// block until exit signal or error
		waitForFinish(ctx, cancel)
	}
}

func handleDaemonModePostCreation(cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider) error {
	apiProvider := apiserver.NewProvider(GetApiServerUrl(), 90, 1*time.Second)

	if err := waitForDaemonModeToBeReady(cancel, kubernetesProvider, apiProvider); err != nil {
		return err
	}
	if err := printDaemonModeTappedPods(apiProvider); err != nil {
		return err
	}

	return nil
}

func printDaemonModeTappedPods(apiProvider *apiserver.Provider) error {
	if healthStatus, err := apiProvider.GetHealthStatus(); err != nil {
		return err
	} else {
		for _, tappedPod := range healthStatus.TapStatus.Pods {
			logger.Log.Infof(uiUtils.Green, fmt.Sprintf("+%s", tappedPod.Name))
		}
	}
	return nil
}

func waitForDaemonModeToBeReady(cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider, apiProvider *apiserver.Provider) error {
	logger.Log.Info("Waiting for mizu to be ready... (may take a few minutes)")
	go startProxyReportErrorIfAny(kubernetesProvider, cancel)

	// TODO: TRA-3903 add a smarter test to see that tapping/pod watching is functioning properly
	if err := apiProvider.TestConnection(); err != nil {
		logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Mizu was not ready in time, for more info check logs at %s", fsUtils.GetLogFilePath()))
		return err
	}
	return nil
}

func startTapperSyncer(ctx context.Context, cancel context.CancelFunc, provider *kubernetes.Provider, targetNamespaces []string, mizuApiFilteringOptions api.TrafficFilteringOptions) error {
	tapperSyncer, err := kubernetes.CreateAndStartMizuTapperSyncer(ctx, provider, kubernetes.TapperSyncerConfig{
		TargetNamespaces:         targetNamespaces,
		PodFilterRegex:           *config.Config.Tap.PodRegex(),
		MizuResourcesNamespace:   config.Config.MizuResourcesNamespace,
		AgentImage:               config.Config.AgentImage,
		TapperResources:          config.Config.Tap.TapperResources,
		ImagePullPolicy:          config.Config.ImagePullPolicy(),
		LogLevel:                 config.Config.LogLevel(),
		IgnoredUserAgents:        config.Config.Tap.IgnoredUserAgents,
		MizuApiFilteringOptions:  mizuApiFilteringOptions,
		MizuServiceAccountExists: state.mizuServiceAccountExists,
	})

	if err != nil {
		return err
	}

	for _, tappedPod := range tapperSyncer.CurrentlyTappedPods {
		logger.Log.Infof(uiUtils.Green, fmt.Sprintf("+%s", tappedPod.Name))
	}

	if len(tapperSyncer.CurrentlyTappedPods) == 0 {
		var suggestionStr string
		if !shared.Contains(targetNamespaces, kubernetes.K8sAllNamespaces) {
			suggestionStr = ". Select a different namespace with -n or tap all namespaces with -A"
		}
		logger.Log.Warningf(uiUtils.Warning, fmt.Sprintf("Did not find any pods matching the regex argument%s", suggestionStr))
	}

	go func() {
		for {
			select {
			case syncerErr, ok := <-tapperSyncer.ErrorOut:
				if !ok {
					logger.Log.Debug("mizuTapperSyncer err channel closed, ending listener loop")
					return
				}
				logger.Log.Errorf(uiUtils.Error, getErrorDisplayTextForK8sTapManagerError(syncerErr))
				cancel()
			case _, ok := <-tapperSyncer.TapPodChangesOut:
				if !ok {
					logger.Log.Debug("mizuTapperSyncer pod changes channel closed, ending listener loop")
					return
				}
				if err := apiProvider.ReportTappedPods(tapperSyncer.CurrentlyTappedPods); err != nil {
					logger.Log.Debugf("[Error] failed update tapped pods %v", err)
				}
			case <-ctx.Done():
				logger.Log.Debug("mizuTapperSyncer event listener loop exiting due to context done")
				return
			}
		}
	}()

	state.tapperSyncer = tapperSyncer

	return nil
}

func getErrorDisplayTextForK8sTapManagerError(err kubernetes.K8sTapManagerError) string {
	switch err.TapManagerReason {
	case kubernetes.TapManagerPodListError:
		return fmt.Sprintf("Failed to update currently tapped pods: %v", err.OriginalError)
	case kubernetes.TapManagerPodWatchError:
		return fmt.Sprintf("Error occured in k8s pod watch: %v", err.OriginalError)
	case kubernetes.TapManagerTapperUpdateError:
		return fmt.Sprintf("Error updating tappers: %v", err.OriginalError)
	default:
		return fmt.Sprintf("Unknown error occured in k8s tap manager: %v", err.OriginalError)
	}
}

func readValidationRules(file string) (string, error) {
	rules, err := shared.DecodeEnforcePolicy(file)
	if err != nil {
		return "", err
	}
	newContent, _ := yaml.Marshal(&rules)
	return string(newContent), nil
}

func createMizuResources(ctx context.Context, cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider, serializedValidationRules string, serializedContract string, serializedMizuConfig string) error {
	if !config.Config.IsNsRestrictedMode() {
		if err := createMizuNamespace(ctx, kubernetesProvider); err != nil {
			return err
		}
	}

	if err := createMizuConfigmap(ctx, kubernetesProvider, serializedValidationRules, serializedContract, serializedMizuConfig); err != nil {
		logger.Log.Warningf(uiUtils.Warning, fmt.Sprintf("Failed to create resources required for policy validation. Mizu will not validate policy rules. error: %v", errormessage.FormatError(err)))
	}

	var err error
	state.mizuServiceAccountExists, err = createRBACIfNecessary(ctx, kubernetesProvider)
	if err != nil {
		if !config.Config.Tap.DaemonMode {
			logger.Log.Warningf(uiUtils.Warning, fmt.Sprintf("Failed to ensure the resources required for IP resolving. Mizu will not resolve target IPs to names. error: %v", errormessage.FormatError(err)))
		}
	}

	var serviceAccountName string
	if state.mizuServiceAccountExists {
		serviceAccountName = kubernetes.ServiceAccountName
	} else {
		serviceAccountName = ""
	}

	opts := &kubernetes.ApiServerOptions{
		Namespace:             config.Config.MizuResourcesNamespace,
		PodName:               kubernetes.ApiServerPodName,
		PodImage:              config.Config.AgentImage,
		ServiceAccountName:    serviceAccountName,
		IsNamespaceRestricted: config.Config.IsNsRestrictedMode(),
		SyncEntriesConfig:     getSyncEntriesConfig(),
		MaxEntriesDBSizeBytes: config.Config.Tap.MaxEntriesDBSizeBytes(),
		Resources:             config.Config.Tap.ApiServerResources,
		ImagePullPolicy:       config.Config.ImagePullPolicy(),
		LogLevel:              config.Config.LogLevel(),
	}

	if config.Config.Tap.DaemonMode {
		if !state.mizuServiceAccountExists {
			defer cleanUpMizuResources(ctx, cancel, kubernetesProvider)
			logger.Log.Fatalf(uiUtils.Red, fmt.Sprintf("Failed to ensure the resources required for mizu to run in daemon mode. cannot proceed. error: %v", errormessage.FormatError(err)))
		}
		if err := createMizuApiServerDeployment(ctx, kubernetesProvider, opts); err != nil {
			return err
		}
	} else {
		if err := createMizuApiServerPod(ctx, kubernetesProvider, opts); err != nil {
			return err
		}
	}

	state.apiServerService, err = kubernetesProvider.CreateService(ctx, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName, kubernetes.ApiServerPodName)
	if err != nil {
		return err
	}
	logger.Log.Debugf("Successfully created service: %s", kubernetes.ApiServerPodName)

	return nil
}

func createMizuConfigmap(ctx context.Context, kubernetesProvider *kubernetes.Provider, serializedValidationRules string, serializedContract string, serializedMizuConfig string) error {
	err := kubernetesProvider.CreateConfigMap(ctx, config.Config.MizuResourcesNamespace, kubernetes.ConfigMapName, serializedValidationRules, serializedContract, serializedMizuConfig)
	return err
}

func createMizuNamespace(ctx context.Context, kubernetesProvider *kubernetes.Provider) error {
	_, err := kubernetesProvider.CreateNamespace(ctx, config.Config.MizuResourcesNamespace)
	return err
}

func createMizuApiServerPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, opts *kubernetes.ApiServerOptions) error {
	pod, err := kubernetesProvider.GetMizuApiServerPodObject(opts, false, "")
	if err != nil {
		return err
	}
	if _, err = kubernetesProvider.CreatePod(ctx, config.Config.MizuResourcesNamespace, pod); err != nil {
		return err
	}
	logger.Log.Debugf("Successfully created API server pod: %s", kubernetes.ApiServerPodName)
	return nil
}

func createMizuApiServerDeployment(ctx context.Context, kubernetesProvider *kubernetes.Provider, opts *kubernetes.ApiServerOptions) error {
	volumeClaimCreated := TryToCreatePersistentVolumeClaim(ctx, kubernetesProvider)

	pod, err := kubernetesProvider.GetMizuApiServerPodObject(opts, volumeClaimCreated, kubernetes.PersistentVolumeClaimName)
	if err != nil {
		return err
	}

	if _, err = kubernetesProvider.CreateDeployment(ctx, config.Config.MizuResourcesNamespace, opts.PodName, pod); err != nil {
		return err
	}
	logger.Log.Debugf("Successfully created API server deployment: %s", kubernetes.ApiServerPodName)
	return nil
}

func TryToCreatePersistentVolumeClaim(ctx context.Context, kubernetesProvider *kubernetes.Provider) bool {
	isDefaultStorageClassAvailable, err := kubernetesProvider.IsDefaultStorageProviderAvailable(ctx)
	if err != nil {
		logger.Log.Debugf("error checking if default storage class exists: %v", err)
	}

	if isDefaultStorageClassAvailable {
		if _, err = kubernetesProvider.CreatePersistentVolumeClaim(ctx, config.Config.MizuResourcesNamespace, kubernetes.PersistentVolumeClaimName, config.Config.Tap.MaxEntriesDBSizeBytes()+mizu.DaemonModePersistentVolumeSizeBufferBytes); err != nil {
			logger.Log.Warningf(uiUtils.Yellow, "An error has occured while creating a persistent volume claim for mizu, this means mizu data will be lost on mizu-api-server pod restart")
			logger.Log.Debugf("error creating persistent volume claim: %v", err)
		} else {
			return true
		}
	} else {
		logger.Log.Warningf(uiUtils.Yellow, "Could not find default volume provider in this cluster, this will mean that mizu's data will be lost on pod restart")
	}
	return false
}

func getMizuApiFilteringOptions() (*api.TrafficFilteringOptions, error) {
	var compiledRegexSlice []*api.SerializableRegexp

	if config.Config.Tap.PlainTextFilterRegexes != nil && len(config.Config.Tap.PlainTextFilterRegexes) > 0 {
		compiledRegexSlice = make([]*api.SerializableRegexp, 0)
		for _, regexStr := range config.Config.Tap.PlainTextFilterRegexes {
			compiledRegex, err := api.CompileRegexToSerializableRegexp(regexStr)
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

func finishMizuExecution(kubernetesProvider *kubernetes.Provider, apiProvider *apiserver.Provider) {
	telemetry.ReportAPICalls(apiProvider)
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
	logger.Log.Infof("\nRemoving mizu resources")

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

	if err := kubernetesProvider.RemoveService(ctx, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName); err != nil {
		resourceDesc := fmt.Sprintf("Service %s in namespace %s", kubernetes.ApiServerPodName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveDaemonSet(ctx, config.Config.MizuResourcesNamespace, kubernetes.TapperDaemonSetName); err != nil {
		resourceDesc := fmt.Sprintf("DaemonSet %s in namespace %s", kubernetes.TapperDaemonSetName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveConfigMap(ctx, config.Config.MizuResourcesNamespace, kubernetes.ConfigMapName); err != nil {
		resourceDesc := fmt.Sprintf("ConfigMap %s in namespace %s", kubernetes.ConfigMapName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveServicAccount(ctx, config.Config.MizuResourcesNamespace, kubernetes.ServiceAccountName); err != nil {
		resourceDesc := fmt.Sprintf("Service Account %s in namespace %s", kubernetes.ServiceAccountName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveRole(ctx, config.Config.MizuResourcesNamespace, kubernetes.RoleName); err != nil {
		resourceDesc := fmt.Sprintf("Role %s in namespace %s", kubernetes.RoleName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemovePod(ctx, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName); err != nil {
		resourceDesc := fmt.Sprintf("Pod %s in namespace %s", kubernetes.ApiServerPodName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	//daemon mode resources
	if err := kubernetesProvider.RemoveRoleBinding(ctx, config.Config.MizuResourcesNamespace, kubernetes.RoleBindingName); err != nil {
		resourceDesc := fmt.Sprintf("RoleBinding %s in namespace %s", kubernetes.RoleBindingName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveDeployment(ctx, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName); err != nil {
		resourceDesc := fmt.Sprintf("Deployment %s in namespace %s", kubernetes.ApiServerPodName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemovePersistentVolumeClaim(ctx, config.Config.MizuResourcesNamespace, kubernetes.PersistentVolumeClaimName); err != nil {
		resourceDesc := fmt.Sprintf("PersistentVolumeClaim %s in namespace %s", kubernetes.PersistentVolumeClaimName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveRole(ctx, config.Config.MizuResourcesNamespace, kubernetes.DaemonRoleName); err != nil {
		resourceDesc := fmt.Sprintf("Role %s in namespace %s", kubernetes.DaemonRoleName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveRoleBinding(ctx, config.Config.MizuResourcesNamespace, kubernetes.DaemonRoleBindingName); err != nil {
		resourceDesc := fmt.Sprintf("RoleBinding %s in namespace %s", kubernetes.DaemonRoleBindingName, config.Config.MizuResourcesNamespace)
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

	if err := kubernetesProvider.RemoveClusterRole(ctx, kubernetes.ClusterRoleName); err != nil {
		resourceDesc := fmt.Sprintf("ClusterRole %s", kubernetes.ClusterRoleName)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveClusterRoleBinding(ctx, kubernetes.ClusterRoleBindingName); err != nil {
		resourceDesc := fmt.Sprintf("ClusterRoleBinding %s", kubernetes.ClusterRoleBindingName)
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

func watchApiServerPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc) {
	podExactRegex := regexp.MustCompile(fmt.Sprintf("^%s$", kubernetes.ApiServerPodName))
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

			logger.Log.Infof("%s removed", kubernetes.ApiServerPodName)
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
				if err := apiProvider.TestConnection(); err != nil {
					logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Couldn't connect to API server, for more info check logs at %s", fsUtils.GetLogFilePath()))
					cancel()
					break
				}

				logger.Log.Infof("Mizu is available at %s", url)
				if !config.Config.HeadlessMode {
					uiUtils.OpenBrowser(url)
				}
				if err := apiProvider.ReportTappedPods(state.tapperSyncer.CurrentlyTappedPods); err != nil {
					logger.Log.Debugf("[Error] failed update tapped pods %v", err)
				}
			}
		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
				continue
			}

			logger.Log.Errorf("[ERROR] Agent creation, watching %v namespace, error: %v", config.Config.MizuResourcesNamespace, err)
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
	podExactRegex := regexp.MustCompile(fmt.Sprintf("^%s.*", kubernetes.TapperDaemonSetName))
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

			logger.Log.Errorf("[Error] Error in mizu tapper watch, err: %v", err)
			cancel()

		case <-ctx.Done():
			logger.Log.Debugf("Watching tapper pod loop, ctx done")
			return
		}
	}
}

func getNamespaces(kubernetesProvider *kubernetes.Provider) []string {
	if config.Config.Tap.AllNamespaces {
		return []string{kubernetes.K8sAllNamespaces}
	} else if len(config.Config.Tap.Namespaces) > 0 {
		return shared.Unique(config.Config.Tap.Namespaces)
	} else {
		currentNamespace, err := kubernetesProvider.CurrentNamespace()
		if err != nil {
			logger.Log.Fatalf(uiUtils.Red, fmt.Sprintf("error getting current namespace: %+v", err))
		}
		return []string{currentNamespace}
	}
}

func createRBACIfNecessary(ctx context.Context, kubernetesProvider *kubernetes.Provider) (bool, error) {
	if !config.Config.IsNsRestrictedMode() {
		if err := kubernetesProvider.CreateMizuRBAC(ctx, config.Config.MizuResourcesNamespace, kubernetes.ServiceAccountName, kubernetes.ClusterRoleName, kubernetes.ClusterRoleBindingName, mizu.RBACVersion); err != nil {
			return false, err
		}
	} else {
		if err := kubernetesProvider.CreateMizuRBACNamespaceRestricted(ctx, config.Config.MizuResourcesNamespace, kubernetes.ServiceAccountName, kubernetes.RoleName, kubernetes.RoleBindingName, mizu.RBACVersion); err != nil {
			return false, err
		}
	}
	if config.Config.Tap.DaemonMode {
		if err := kubernetesProvider.CreateDaemonsetRBAC(ctx, config.Config.MizuResourcesNamespace, kubernetes.ServiceAccountName, kubernetes.DaemonRoleName, kubernetes.DaemonRoleBindingName, mizu.RBACVersion); err != nil {
			return false, err
		}
	}
	return true, nil
}
