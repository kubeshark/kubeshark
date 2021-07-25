package cmd

import (
	"context"
	"fmt"
	"github.com/romana/rlog"
	"github.com/up9inc/mizu/cli/kubernetes"
	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/debounce"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
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

const (
	cleanupTimeout     = time.Minute
	updateTappersDelay = 5 * time.Second
)

type tapCmdBL struct {
	apiServerService         *core.Service
	currentlyTappedPods      []core.Pod
	flags                    *MizuTapOptions
	mizuServiceAccountExists bool
	resourcesNamespace       string
}

func NewtapCmdBL(flags *MizuTapOptions) *tapCmdBL {
	var (
		resourcesNamespace string
	)
	if flags.MizuNamespace != "" {
		resourcesNamespace = flags.MizuNamespace
	} else {
		resourcesNamespace = mizu.ResourcesDefaultNamespace
	}

	return &tapCmdBL{
		flags: flags,
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
			fmt.Printf(mizu.Red, "Couldn't find the kube config file, or file is empty. Try adding '--kube-config=<path to kube config file>'\n")
			return
		}
		if clientcmd.IsConfigurationInvalid(err) {
			fmt.Printf(mizu.Red, "Invalid kube config file. Try using a different config with '--kube-config=<path to kube config file>'\n")
			return
		}
	}

	defer bl.cleanUpMizuResources(kubernetesProvider)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	targetNamespace := bl.getNamespace(kubernetesProvider)
	if matchingPods, err := kubernetesProvider.GetAllPodsMatchingRegex(ctx, podRegexQuery, targetNamespace); err != nil {
		fmt.Printf("Error listing pods: %v", err)
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
	fmt.Printf("Tapping pods in %s\n", namespacesStr)

	if len(bl.currentlyTappedPods) == 0 {
		var suggestionStr string
		if targetNamespace != mizu.K8sAllNamespaces {
			suggestionStr = "\nSelect a different namespace with -n or tap all namespaces with -A"
		}
		fmt.Printf("Did not find any pods matching the regex argument%s\n", suggestionStr)
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
	if err := createMizuNamespace(ctx, kubernetesProvider, bl.resourcesNamespace); err != nil {
		return err
	}

	if err := bl.createMizuApiServer(ctx, kubernetesProvider, mizuApiFilteringOptions); err != nil {
		return err
	}

	if err := bl.updateMizuTappers(ctx, kubernetesProvider, nodeToTappedPodIPMap); err != nil {
		return err
	}

	return nil
}

func createMizuNamespace(ctx context.Context, kubernetesProvider *kubernetes.Provider, namespace string) error {
	_, err := kubernetesProvider.CreateNamespace(ctx, namespace)
	if err != nil {
		fmt.Printf("Error creating Namespace %s: %v\n", namespace, err)
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
	_, err = kubernetesProvider.CreateMizuApiServerPod(ctx, bl.resourcesNamespace, mizu.ApiServerPodName, bl.flags.MizuImage, serviceAccountName, mizuApiFilteringOptions, bl.flags.MaxEntriesDBSizeBytes)
	if err != nil {
		fmt.Printf("Error creating mizu %s pod: %v\n", mizu.ApiServerPodName, err)
		return err
	}

	bl.apiServerService, err = kubernetesProvider.CreateService(ctx, bl.resourcesNamespace, mizu.ApiServerPodName, mizu.ApiServerPodName)
	if err != nil {
		fmt.Printf("Error creating mizu %s service: %v\n", mizu.ApiServerPodName,  err)
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
				fmt.Printf("Regex %s is invalid: %v", regexStr, err)
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
			fmt.Printf("Error creating mizu tapper daemonset: %v\n", err)
			return err
		}
	} else {
		if err := kubernetesProvider.RemoveDaemonSet(ctx, bl.resourcesNamespace, mizu.TapperDaemonSetName); err != nil {
			fmt.Printf("Error deleting mizu tapper daemonset: %v\n", err)
			return err
		}
	}

	return nil
}

func (bl *tapCmdBL) cleanUpMizuResources(kubernetesProvider *kubernetes.Provider) {
	fmt.Printf("\nRemoving mizu resources\n")

	removalCtx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()

	if err := kubernetesProvider.RemoveNamespace(removalCtx, bl.resourcesNamespace); err != nil {
		fmt.Printf("Error removing Namespace %s: %s (%v,%+v)\n", bl.resourcesNamespace, err, err, err)
		return
	}

	if bl.mizuServiceAccountExists {
		if err := kubernetesProvider.RemoveNonNamespacedResources(removalCtx, mizu.ClusterRoleName, mizu.ClusterRoleBindingName); err != nil {
			fmt.Printf("Error removing non-namespaced resources: %s (%v,%+v)\n", err, err, err)
			return
		}
	}

	// Call cancel if a terminating signal was received. Allows user to skip the wait.
	go func() {
		waitForFinish(removalCtx, cancel)
	}()

	if err := kubernetesProvider.WaitUtilNamespaceDeleted(removalCtx, bl.resourcesNamespace); err != nil {
		switch {
		case removalCtx.Err() == context.Canceled:
			// Do nothing. User interrupted the wait.
		case err == wait.ErrWaitTimeout:
			fmt.Printf("Timeout while removing Namespace %s\n", bl.resourcesNamespace)
		default:
			fmt.Printf("Error while waiting for Namespace %s to be deleted: %s (%v,%+v)\n", bl.resourcesNamespace, err, err, err)
		}
	}
}

func (bl *tapCmdBL) watchPodsForTapping(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc, podRegex *regexp.Regexp) {
	targetNamespace := bl.getNamespace(kubernetesProvider)

	added, modified, removed, errorChan := kubernetes.FilteredWatch(ctx, kubernetesProvider.GetPodWatcher(ctx, targetNamespace), podRegex)

	restartTappers := func() {
		if matchingPods, err := kubernetesProvider.GetAllPodsMatchingRegex(ctx, podRegex, targetNamespace); err != nil {
			fmt.Printf("Error getting pods by regex: %s (%v,%+v)\n", err, err, err)
			cancel()
		} else {
			bl.currentlyTappedPods = matchingPods
		}

		nodeToTappedPodIPMap, err := getNodeHostToTappedPodIpsMap(bl.currentlyTappedPods)
		if err != nil {
			fmt.Printf("Error building node to ips map: %s (%v,%+v)\n", err, err, err)
			cancel()
		}

		if err := bl.updateMizuTappers(ctx, kubernetesProvider, nodeToTappedPodIPMap); err != nil {
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
			fmt.Printf("%s removed\n", mizu.ApiServerPodName)
			cancel()
			return
		case modifiedPod := <-modified:
			if modifiedPod.Status.Phase == "Running" && !isPodReady {
				isPodReady = true
				go func() {
					err := kubernetes.StartProxy(kubernetesProvider, bl.flags.GuiPort, bl.resourcesNamespace, mizu.ApiServerPodName)
					if err != nil {
						fmt.Printf("Error occured while running k8s proxy %v\n", err)
						cancel()
					}
				}()
				mizuProxiedUrl := kubernetes.GetMizuApiServerProxiedHostAndPath(bl.flags.GuiPort)
				fmt.Printf("Mizu is available at http://%s\n", mizuProxiedUrl)

				time.Sleep(time.Second * 5) // Waiting to be sure the proxy is ready
				if bl.flags.Analysis {
					urlPath := fmt.Sprintf("http://%s/api/uploadEntries?dest=%s&interval=%v", mizuProxiedUrl, url.QueryEscape(bl.flags.AnalysisDestination), bl.flags.SleepIntervalSec)
					u, err := url.ParseRequestURI(urlPath)

					if err != nil {
						log.Fatal(fmt.Sprintf("Failed parsing the URL %v\n", err))
					}
					rlog.Debugf("Sending get request to %v\n", u.String())
					if response, err := http.Get(u.String()); err != nil || response.StatusCode != 200 {
						fmt.Printf("error sending upload entries req, status code: %v, err: %v\n", response.StatusCode, err)
					} else {
						fmt.Printf(mizu.Purple, "Traffic is uploading to UP9 for further analsys")
						fmt.Println()
					}
				}
			}

		case <-timeAfter:
			if !isPodReady {
				fmt.Printf("error: %s pod was not ready in time", mizu.ApiServerPodName)
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
		fmt.Printf("warning: could not ensure mizu rbac resources exist %v\n", err)
		return false
	}
	if !mizuRBACExists {
		err := kubernetesProvider.CreateMizuRBAC(ctx, bl.resourcesNamespace, mizu.ServiceAccountName, mizu.ClusterRoleName, mizu.ClusterRoleBindingName, mizu.RBACVersion)
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

func (bl *tapCmdBL) syncApiStatus(ctx context.Context, cancel context.CancelFunc) {
	controlSocketStr := fmt.Sprintf("ws://%s/ws", kubernetes.GetMizuApiServerProxiedHostAndPath(bl.flags.GuiPort))
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
			err = controlSocket.SendNewTappedPodsListMessage(bl.currentlyTappedPods)
			if err != nil {
				rlog.Debugf("error Sending message via control socket %v, error: %s\n", controlSocketStr, err)
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
