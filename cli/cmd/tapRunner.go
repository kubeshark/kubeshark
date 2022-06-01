package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/up9inc/mizu/cli/resources"
	"github.com/up9inc/mizu/cli/telemetry"
	"github.com/up9inc/mizu/cli/utils"

	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
	core "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/up9inc/mizu/cli/apiserver"
	"github.com/up9inc/mizu/cli/cmd/goUtils"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/errormessage"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/kubernetes"
	"github.com/up9inc/mizu/tap/api"
)

const cleanupTimeout = time.Minute

type tapState struct {
	startTime                time.Time
	targetNamespaces         []string
	mizuServiceAccountExists bool
}

var state tapState
var apiProvider *apiserver.Provider

func RunMizuTap() {
	state.startTime = time.Now()

	apiProvider = apiserver.NewProvider(GetApiServerUrl(config.Config.Tap.GuiPort), apiserver.DefaultRetries, apiserver.DefaultTimeout)

	var err error
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

	state.targetNamespaces = getNamespaces(kubernetesProvider)

	mizuAgentConfig := getTapMizuAgentConfig()
	serializedMizuConfig, err := getSerializedMizuAgentConfig(mizuAgentConfig)
	if err != nil {
		logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error serializing mizu config: %v", errormessage.FormatError(err)))
		return
	}

	if config.Config.IsNsRestrictedMode() {
		if len(state.targetNamespaces) != 1 || !shared.Contains(state.targetNamespaces, config.Config.MizuResourcesNamespace) {
			logger.Log.Errorf("Not supported mode. Mizu can't resolve IPs in other namespaces when running in namespace restricted mode.\n"+
				"You can use the same namespace for --%s and --%s", configStructs.NamespacesTapName, config.MizuResourcesNamespaceConfigName)
			return
		}
	}

	var namespacesStr string
	if !shared.Contains(state.targetNamespaces, kubernetes.K8sAllNamespaces) {
		namespacesStr = fmt.Sprintf("namespaces \"%s\"", strings.Join(state.targetNamespaces, "\", \""))
	} else {
		namespacesStr = "all namespaces"
	}

	logger.Log.Infof("Tapping pods in %s", namespacesStr)

	if err := printTappedPodsPreview(ctx, kubernetesProvider, state.targetNamespaces); err != nil {
		logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error listing pods: %v", errormessage.FormatError(err)))
	}

	if config.Config.Tap.DryRun {
		return
	}

	logger.Log.Infof("Waiting for Mizu Agent to start...")
	if state.mizuServiceAccountExists, err = resources.CreateTapMizuResources(ctx, kubernetesProvider, serializedValidationRules, serializedContract, serializedMizuConfig, config.Config.IsNsRestrictedMode(), config.Config.MizuResourcesNamespace, config.Config.AgentImage, config.Config.Tap.MaxEntriesDBSizeBytes(), config.Config.Tap.ApiServerResources, config.Config.ImagePullPolicy(), config.Config.LogLevel(), config.Config.Tap.Profiler); err != nil {
		var statusError *k8serrors.StatusError
		if errors.As(err, &statusError) && (statusError.ErrStatus.Reason == metav1.StatusReasonAlreadyExists) {
			logger.Log.Info("Mizu is already running in this namespace, change the `mizu-resources-namespace` configuration or run `mizu clean` to remove the currently running Mizu instance")
		} else {
			defer resources.CleanUpMizuResources(ctx, cancel, kubernetesProvider, config.Config.IsNsRestrictedMode(), config.Config.MizuResourcesNamespace)
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error creating resources: %v", errormessage.FormatError(err)))
		}

		return
	}

	defer finishTapExecution(kubernetesProvider)

	go goUtils.HandleExcWrapper(watchApiServerEvents, ctx, kubernetesProvider, cancel)
	go goUtils.HandleExcWrapper(watchApiServerPod, ctx, kubernetesProvider, cancel)

	// block until exit signal or error
	utils.WaitForFinish(ctx, cancel)
}

func finishTapExecution(kubernetesProvider *kubernetes.Provider) {
	telemetry.ReportTapTelemetry(apiProvider, config.Config.Tap, state.startTime)

	finishMizuExecution(kubernetesProvider, config.Config.IsNsRestrictedMode(), config.Config.MizuResourcesNamespace)
}

func getTapMizuAgentConfig() *shared.MizuAgentConfig {
	mizuAgentConfig := shared.MizuAgentConfig{
		MaxDBSizeBytes:         config.Config.Tap.MaxEntriesDBSizeBytes(),
		InsertionFilter:        config.Config.Tap.GetInsertionFilter(),
		AgentImage:             config.Config.AgentImage,
		PullPolicy:             config.Config.ImagePullPolicyStr,
		LogLevel:               config.Config.LogLevel(),
		TapperResources:        config.Config.Tap.TapperResources,
		MizuResourcesNamespace: config.Config.MizuResourcesNamespace,
		AgentDatabasePath:      shared.DataDirPath,
		ServiceMap:             config.Config.ServiceMap,
		OAS:                    config.Config.OAS,
		Telemetry:              config.Config.Telemetry,
	}

	return &mizuAgentConfig
}

/*
this function is a bit problematic as it might be detached from the actual pods the mizu api server will tap.
The alternative would be to wait for api server to be ready and then query it for the pods it listens to, this has
the arguably worse drawback of taking a relatively very long time before the user sees which pods are targeted, if any.
*/
func printTappedPodsPreview(ctx context.Context, kubernetesProvider *kubernetes.Provider, namespaces []string) error {
	if matchingPods, err := kubernetesProvider.ListAllRunningPodsMatchingRegex(ctx, config.Config.Tap.PodRegex(), namespaces); err != nil {
		return err
	} else {
		if len(matchingPods) == 0 {
			printNoPodsFoundSuggestion(namespaces)
		}
		for _, tappedPod := range matchingPods {
			logger.Log.Infof(uiUtils.Green, fmt.Sprintf("+%s", tappedPod.Name))
		}
		return nil
	}
}

func startTapperSyncer(ctx context.Context, cancel context.CancelFunc, provider *kubernetes.Provider, targetNamespaces []string, mizuApiFilteringOptions api.TrafficFilteringOptions, startTime time.Time) error {
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
		ServiceMesh:              config.Config.Tap.ServiceMesh,
		Tls:                      config.Config.Tap.Tls,
	}, startTime)

	if err != nil {
		return err
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
			case tapperStatus, ok := <-tapperSyncer.TapperStatusChangedOut:
				if !ok {
					logger.Log.Debug("mizuTapperSyncer tapper status changed channel closed, ending listener loop")
					return
				}
				if err := apiProvider.ReportTapperStatus(tapperStatus); err != nil {
					logger.Log.Debugf("[Error] failed update tapper status %v", err)
				}
			case <-ctx.Done():
				logger.Log.Debug("mizuTapperSyncer event listener loop exiting due to context done")
				return
			}
		}
	}()

	return nil
}

func printNoPodsFoundSuggestion(targetNamespaces []string) {
	var suggestionStr string
	if !shared.Contains(targetNamespaces, kubernetes.K8sAllNamespaces) {
		suggestionStr = ". You can also try selecting a different namespace with -n or tap all namespaces with -A"
	}
	logger.Log.Warningf(uiUtils.Warning, fmt.Sprintf("Did not find any currently running pods that match the regex argument, mizu will automatically tap matching pods if any are created later%s", suggestionStr))
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
		EnableRedaction:        config.Config.Tap.EnableRedaction,
	}, nil
}

func watchApiServerPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc) {
	podExactRegex := regexp.MustCompile(fmt.Sprintf("^%s$", kubernetes.ApiServerPodName))
	podWatchHelper := kubernetes.NewPodWatchHelper(kubernetesProvider, podExactRegex)
	eventChan, errorChan := kubernetes.FilteredWatch(ctx, podWatchHelper, []string{config.Config.MizuResourcesNamespace}, podWatchHelper)
	isPodReady := false

	apiServerTimeoutSec := config.GetIntEnvConfig(config.ApiServerTimeoutSec, 120)
	timeAfter := time.After(time.Duration(apiServerTimeoutSec) * time.Second)
	for {
		select {
		case wEvent, ok := <-eventChan:
			if !ok {
				eventChan = nil
				continue
			}

			switch wEvent.Type {
			case kubernetes.EventAdded:
				logger.Log.Debugf("Watching API Server pod loop, added")
			case kubernetes.EventDeleted:
				logger.Log.Infof("%s removed", kubernetes.ApiServerPodName)
				cancel()
				return
			case kubernetes.EventModified:
				modifiedPod, err := wEvent.ToPod()
				if err != nil {
					logger.Log.Errorf(uiUtils.Error, err)
					cancel()
					continue
				}

				logger.Log.Debugf("Watching API Server pod loop, modified: %v, containers statuses: %v", modifiedPod.Status.Phase, modifiedPod.Status.ContainerStatuses)

				if modifiedPod.Status.Phase == core.PodRunning && !isPodReady {
					isPodReady = true
					postApiServerStarted(ctx, kubernetesProvider, cancel)
				}
			case kubernetes.EventBookmark:
				break
			case kubernetes.EventError:
				break
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

func watchApiServerEvents(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc) {
	podExactRegex := regexp.MustCompile(fmt.Sprintf("^%s", kubernetes.ApiServerPodName))
	eventWatchHelper := kubernetes.NewEventWatchHelper(kubernetesProvider, podExactRegex, "pod")
	eventChan, errorChan := kubernetes.FilteredWatch(ctx, eventWatchHelper, []string{config.Config.MizuResourcesNamespace}, eventWatchHelper)
	for {
		select {
		case wEvent, ok := <-eventChan:
			if !ok {
				eventChan = nil
				continue
			}

			event, err := wEvent.ToEvent()
			if err != nil {
				logger.Log.Debugf("[ERROR] parsing Mizu resource event: %+v", err)
				continue
			}

			if state.startTime.After(event.CreationTimestamp.Time) {
				continue
			}

			logger.Log.Debugf(
				fmt.Sprintf("Watching API server events loop, event %s, time: %v, resource: %s (%s), reason: %s, note: %s",
					event.Name,
					event.CreationTimestamp.Time,
					event.Regarding.Name,
					event.Regarding.Kind,
					event.Reason,
					event.Note))

			switch event.Reason {
			case "FailedScheduling", "Failed":
				logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Mizu API Server status: %s - %s", event.Reason, event.Note))
				cancel()

			}
		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
				continue
			}

			logger.Log.Debugf("[Error] Watching API server events loop, error: %+v", err)
		case <-ctx.Done():
			logger.Log.Debugf("Watching API server events loop, ctx done")
			return
		}
	}
}

func postApiServerStarted(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc) {
	startProxyReportErrorIfAny(kubernetesProvider, ctx, cancel, config.Config.Tap.GuiPort)

	options, _ := getMizuApiFilteringOptions()
	if err := startTapperSyncer(ctx, cancel, kubernetesProvider, state.targetNamespaces, *options, state.startTime); err != nil {
		logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error starting mizu tapper syncer: %v", errormessage.FormatError(err)))
		cancel()
	}

	url := GetApiServerUrl(config.Config.Tap.GuiPort)
	logger.Log.Infof("Mizu is available at %s", url)
	if !config.Config.HeadlessMode {
		uiUtils.OpenBrowser(url)
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
