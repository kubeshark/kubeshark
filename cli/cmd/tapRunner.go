package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/up9inc/mizu/cli/kubernetes"
	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/debounce"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	cleanupTimeout     = time.Minute
	updateTappersDelay = 5 * time.Second
)

type tapCmdBL struct {
	apiServerService         *core.Service
	currentlyTappedPods      []core.Pod
	flags                    *MizuTapOptions
	mizuServiceAccountExists bool
	isOwnNamespace           bool
	resourcesNamespace       string
}

func NewTapCmdBL(flags *MizuTapOptions) *tapCmdBL {
	var (
		isOwnNamespace bool
		resourcesNamespace string
	)
	if flags.MizuNamespace != "" {
		isOwnNamespace = false
		resourcesNamespace = flags.MizuNamespace
	} else {
		isOwnNamespace = true
		resourcesNamespace = mizu.ResourcesDefaultNamespace
	}

	return &tapCmdBL{
		flags: flags,
		isOwnNamespace: isOwnNamespace,
		resourcesNamespace: resourcesNamespace,
	}
}

func (bl *tapCmdBL) RunMizuTap(podRegexQuery *regexp.Regexp) {
	mizuApiFilteringOptions, err := bl.getMizuApiFilteringOptions()
	if err != nil {
		return
	}

	kubernetesProvider, err := kubernetes.NewProvider(bl.flags.KubeConfigPath)
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

	defer bl.cleanUpMizuResources(kubernetesProvider)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	targetNamespace := bl.getNamespace(kubernetesProvider)
	if matchingPods, err := kubernetesProvider.GetAllPodsMatchingRegex(ctx, podRegexQuery, targetNamespace); err != nil {
		mizu.Log.Infof("Error listing pods: %v", err)
		return
	} else {
		bl.currentlyTappedPods = matchingPods
	}

	var namespacesStr string
	if targetNamespace != mizu.K8sAllNamespaces {
		namespacesStr = fmt.Sprintf("namespace \"%s\"", targetNamespace)
	} else {
		namespacesStr = "all namespaces"
	}
	mizu.Log.Infof("Tapping pods in %s", namespacesStr)

	if len(bl.currentlyTappedPods) == 0 {
		var suggestionStr string
		if targetNamespace != mizu.K8sAllNamespaces {
			suggestionStr = "\nSelect a different namespace with -n or tap all namespaces with -A"
		}
		mizu.Log.Infof("Did not find any pods matching the regex argument%s", suggestionStr)
	}

	nodeToTappedPodIPMap, err := getNodeHostToTappedPodIpsMap(bl.currentlyTappedPods)
	if err != nil {
		return
	}

	if err := bl.createMizuResources(ctx, kubernetesProvider, nodeToTappedPodIPMap, mizuApiFilteringOptions); err != nil {
		return
	}

	mizu.CheckNewerVersion()
	go bl.portForwardApiPod(ctx, kubernetesProvider, cancel) // TODO convert this to job for built in pod ttl or have the running app handle this
	go bl.watchPodsForTapping(ctx, kubernetesProvider, cancel, podRegexQuery)
	go bl.syncApiStatus(ctx, cancel)

	//block until exit signal or error
	waitForFinish(ctx, cancel)
}

func (bl *tapCmdBL) createMizuResources(ctx context.Context, kubernetesProvider *kubernetes.Provider, nodeToTappedPodIPMap map[string][]string, mizuApiFilteringOptions *shared.TrafficFilteringOptions) error {
	if bl.isOwnNamespace {
		if err := bl.createMizuNamespace(ctx, kubernetesProvider); err != nil {
			return err
		}
	}

	if err := bl.createMizuApiServer(ctx, kubernetesProvider, mizuApiFilteringOptions); err != nil {
		return err
	}

	if err := bl.updateMizuTappers(ctx, kubernetesProvider, nodeToTappedPodIPMap); err != nil {
		return err
	}

	return nil
}

func (bl *tapCmdBL) createMizuNamespace(ctx context.Context, kubernetesProvider *kubernetes.Provider) error {
	_, err := kubernetesProvider.CreateNamespace(ctx, bl.resourcesNamespace)
	if err != nil {
		mizu.Log.Infof("Error creating Namespace %s: %v", bl.resourcesNamespace, err)
	}

	return err
}

func (bl *tapCmdBL) createMizuApiServer(ctx context.Context, kubernetesProvider *kubernetes.Provider, mizuApiFilteringOptions *shared.TrafficFilteringOptions) error {
	var err error

	bl.mizuServiceAccountExists = bl.createRBACIfNecessary(ctx, kubernetesProvider)
	var serviceAccountName string
	if bl.mizuServiceAccountExists {
		serviceAccountName = mizu.ServiceAccountName
	} else {
		serviceAccountName = ""
	}

	opts := &kubernetes.ApiServerOptions{
		Namespace: bl.resourcesNamespace,
		PodName: mizu.ApiServerPodName,
		PodImage: bl.flags.MizuImage,
		ServiceAccountName: serviceAccountName,
		IsNamespaceRestricted: !bl.isOwnNamespace,
		MizuApiFilteringOptions: mizuApiFilteringOptions,
		MaxEntriesDBSizeBytes: bl.flags.MaxEntriesDBSizeBytes,
	}
	_, err = kubernetesProvider.CreateMizuApiServerPod(ctx, opts)
	if err != nil {
		mizu.Log.Infof("Error creating mizu %s pod: %v", mizu.ApiServerPodName, err)
		return err
	}

	bl.apiServerService, err = kubernetesProvider.CreateService(ctx, bl.resourcesNamespace, mizu.ApiServerPodName, mizu.ApiServerPodName)
	if err != nil {
		mizu.Log.Infof("Error creating mizu %s service: %v", mizu.ApiServerPodName, err)
		return err
	}

	return nil
}

func (bl *tapCmdBL) getMizuApiFilteringOptions() (*shared.TrafficFilteringOptions, error) {
	var compiledRegexSlice []*shared.SerializableRegexp

	if bl.flags.PlainTextFilterRegexes != nil && len(bl.flags.PlainTextFilterRegexes) > 0 {
		compiledRegexSlice = make([]*shared.SerializableRegexp, 0)
		for _, regexStr := range bl.flags.PlainTextFilterRegexes {
			compiledRegex, err := shared.CompileRegexToSerializableRegexp(regexStr)
			if err != nil {
				mizu.Log.Infof("Regex %s is invalid: %v", regexStr, err)
				return nil, err
			}
			compiledRegexSlice = append(compiledRegexSlice, compiledRegex)
		}
	}

	return &shared.TrafficFilteringOptions{PlainTextMaskingRegexes: compiledRegexSlice, HideHealthChecks: bl.flags.HideHealthChecks, DisableRedaction: bl.flags.DisableRedaction}, nil
}

func (bl *tapCmdBL) updateMizuTappers(ctx context.Context, kubernetesProvider *kubernetes.Provider, nodeToTappedPodIPMap map[string][]string) error {
	if len(nodeToTappedPodIPMap) > 0 {
		var serviceAccountName string
		if bl.mizuServiceAccountExists {
			serviceAccountName = mizu.ServiceAccountName
		} else {
			serviceAccountName = ""
		}

		if err := kubernetesProvider.ApplyMizuTapperDaemonSet(
			ctx,
			bl.resourcesNamespace,
			mizu.TapperDaemonSetName,
			bl.flags.MizuImage,
			mizu.TapperPodName,
			fmt.Sprintf("%s.%s.svc.cluster.local", bl.apiServerService.Name, bl.apiServerService.Namespace),
			nodeToTappedPodIPMap,
			serviceAccountName,
			bl.flags.TapOutgoing,
		); err != nil {
			mizu.Log.Infof("Error creating mizu tapper daemonset: %v", err)
			return err
		}
	} else {
		if err := kubernetesProvider.RemoveDaemonSet(ctx, bl.resourcesNamespace, mizu.TapperDaemonSetName); err != nil {
			mizu.Log.Infof("Error deleting mizu tapper daemonset: %v", err)
			return err
		}
	}

	return nil
}

func (bl *tapCmdBL) cleanUpMizuResources(kubernetesProvider *kubernetes.Provider) {
	mizu.Log.Infof("\nRemoving mizu resources\n")

	removalCtx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()

	if bl.isOwnNamespace {
		if err := kubernetesProvider.RemoveNamespace(removalCtx, bl.resourcesNamespace); err != nil {
			mizu.Log.Infof("Error removing Namespace %s: %s (%v,%+v)", bl.resourcesNamespace, err, err, err)
			return
		}
	} else {
		if err := kubernetesProvider.RemovePod(removalCtx, bl.resourcesNamespace, mizu.ApiServerPodName); err != nil {
			mizu.Log.Infof("Error removing Pod %s in namespace %s: %s (%v,%+v)", mizu.ApiServerPodName, bl.resourcesNamespace, err, err, err)
		}

		if err := kubernetesProvider.RemoveService(removalCtx, bl.resourcesNamespace, mizu.ApiServerPodName); err != nil {
			mizu.Log.Infof("Error removing Service %s in namespace %s: %s (%v,%+v)", mizu.ApiServerPodName, bl.resourcesNamespace, err, err, err)
		}

		if err := kubernetesProvider.RemoveDaemonSet(removalCtx, bl.resourcesNamespace, mizu.TapperDaemonSetName); err != nil {
			mizu.Log.Infof("Error removing DaemonSet %s in namespace %s: %s (%v,%+v)", mizu.TapperDaemonSetName, bl.resourcesNamespace, err, err, err)
		}
	}

	if bl.mizuServiceAccountExists {
		if bl.isOwnNamespace {
			if err := kubernetesProvider.RemoveNonNamespacedResources(removalCtx, mizu.ClusterRoleName, mizu.ClusterRoleBindingName); err != nil {
				mizu.Log.Infof("Error removing non-namespaced resources: %s (%v,%+v)", err, err, err)
				return
			}
		} else {
			if err := kubernetesProvider.RemoveServicAccount(removalCtx, bl.resourcesNamespace, mizu.ServiceAccountName); err != nil {
				mizu.Log.Infof("Error removing Service Account %s in namespace %s: %s (%v,%+v)", mizu.ServiceAccountName, bl.resourcesNamespace, err, err, err)
				return
			}

			if err := kubernetesProvider.RemoveRole(removalCtx, bl.resourcesNamespace, mizu.RoleName); err != nil {
				mizu.Log.Infof("Error removing Role %s in namespace %s: %s (%v,%+v)", mizu.RoleName, bl.resourcesNamespace, err, err, err)
			}

			if err := kubernetesProvider.RemoveRoleBinding(removalCtx, bl.resourcesNamespace, mizu.RoleBindingName); err != nil {
				mizu.Log.Infof("Error removing RoleBinding %s in namespace %s: %s (%v,%+v)", mizu.RoleBindingName, bl.resourcesNamespace, err, err, err)
			}
		}
	}

	if bl.isOwnNamespace {
		bl.waitUntilNamespaceDeleted(removalCtx, cancel, kubernetesProvider)
	}
}

func (bl *tapCmdBL) waitUntilNamespaceDeleted(ctx context.Context, cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider) {
	// Call cancel if a terminating signal was received. Allows user to skip the wait.
	go func() {
		waitForFinish(ctx, cancel)
	}()

	if err := kubernetesProvider.WaitUtilNamespaceDeleted(ctx, bl.resourcesNamespace); err != nil {
		switch {
		case ctx.Err() == context.Canceled:
			// Do nothing. User interrupted the wait.
		case err == wait.ErrWaitTimeout:
			mizu.Log.Infof("Timeout while removing Namespace %s", bl.resourcesNamespace)
		default:
			mizu.Log.Infof("Error while waiting for Namespace %s to be deleted: %s (%v,%+v)", bl.resourcesNamespace, err, err, err)
		}
	}
}

func (bl *tapCmdBL) watchPodsForTapping(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc, podRegex *regexp.Regexp) {
	targetNamespace := bl.getNamespace(kubernetesProvider)

	added, modified, removed, errorChan := kubernetes.FilteredWatch(ctx, kubernetesProvider.GetPodWatcher(ctx, targetNamespace), podRegex)

	restartTappers := func() {
		if matchingPods, err := kubernetesProvider.GetAllPodsMatchingRegex(ctx, podRegex, targetNamespace); err != nil {
			mizu.Log.Infof("Error getting pods by regex: %s (%v,%+v)", err, err, err)
			cancel()
		} else {
			bl.currentlyTappedPods = matchingPods
		}

		nodeToTappedPodIPMap, err := getNodeHostToTappedPodIpsMap(bl.currentlyTappedPods)
		if err != nil {
			mizu.Log.Infof("Error building node to ips map: %s (%v,%+v)", err, err, err)
			cancel()
		}

		if err := bl.updateMizuTappers(ctx, kubernetesProvider, nodeToTappedPodIPMap); err != nil {
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

func (bl *tapCmdBL) portForwardApiPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc) {
	podExactRegex := regexp.MustCompile(fmt.Sprintf("^%s$", mizu.ApiServerPodName))
	added, modified, removed, errorChan := kubernetes.FilteredWatch(ctx, kubernetesProvider.GetPodWatcher(ctx, bl.resourcesNamespace), podExactRegex)
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
					err := kubernetes.StartProxy(kubernetesProvider, bl.flags.GuiPort, bl.resourcesNamespace, mizu.ApiServerPodName)
					if err != nil {
						mizu.Log.Infof("Error occured while running k8s proxy %v", err)
						cancel()
					}
				}()
				mizuProxiedUrl := kubernetes.GetMizuApiServerProxiedHostAndPath(bl.flags.GuiPort)
				mizu.Log.Infof("Mizu is available at http://%s", mizuProxiedUrl)

				time.Sleep(time.Second * 5) // Waiting to be sure the proxy is ready
				if bl.flags.Analysis {
					urlPath := fmt.Sprintf("http://%s/api/uploadEntries?dest=%s&interval=%v", mizuProxiedUrl, url.QueryEscape(bl.flags.AnalysisDestination), bl.flags.SleepIntervalSec)
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

func (bl *tapCmdBL) createRBACIfNecessary(ctx context.Context, kubernetesProvider *kubernetes.Provider) bool {
	mizuRBACExists, err := kubernetesProvider.DoesServiceAccountExist(ctx, bl.resourcesNamespace, mizu.ServiceAccountName)
	if err != nil {
		mizu.Log.Infof("warning: could not ensure mizu rbac resources exist %v", err)
		return false
	}
	if !mizuRBACExists {
		if bl.isOwnNamespace {
			err := kubernetesProvider.CreateMizuRBAC(ctx, bl.resourcesNamespace, mizu.ServiceAccountName, mizu.ClusterRoleName, mizu.ClusterRoleBindingName, mizu.RBACVersion)
			if err != nil {
				mizu.Log.Infof("warning: could not create mizu rbac resources %v", err)
				return false
			}
		} else {
			err := kubernetesProvider.CreateMizuRBACNamespaceRestricted(ctx, bl.resourcesNamespace, mizu.ServiceAccountName, mizu.RoleName, mizu.RoleBindingName, mizu.RBACVersion)
			if err != nil {
				mizu.Log.Infof("warning: could not create mizu rbac resources %v", err)
				return false
			}
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

func (bl *tapCmdBL) syncApiStatus(ctx context.Context, cancel context.CancelFunc) {
	controlSocketStr := fmt.Sprintf("ws://%s/ws", kubernetes.GetMizuApiServerProxiedHostAndPath(bl.flags.GuiPort))
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
			err = controlSocket.SendNewTappedPodsListMessage(bl.currentlyTappedPods)
			if err != nil {
				mizu.Log.Debugf("error Sending message via control socket %v, error: %s", controlSocketStr, err)
			}
			time.Sleep(10 * time.Second)
		}
	}

}

func (bl *tapCmdBL) getNamespace(kubernetesProvider *kubernetes.Provider) string {
	if bl.flags.AllNamespaces {
		return mizu.K8sAllNamespaces
	} else if len(bl.flags.Namespace) > 0 {
		return bl.flags.Namespace
	} else {
		return kubernetesProvider.CurrentNamespace()
	}
}
