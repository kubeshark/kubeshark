package kubernetes

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/misc"
	"github.com/kubeshark/kubeshark/semver"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/rs/zerolog/log"
	"github.com/tanqiangyes/grep-go/reader"
	core "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Provider struct {
	clientSet        *kubernetes.Clientset
	kubernetesConfig clientcmd.ClientConfig
	clientConfig     rest.Config
	managedBy        string
	createdBy        string
}

func NewProvider(kubeConfigPath string, contextName string) (*Provider, error) {
	kubernetesConfig := loadKubernetesConfiguration(kubeConfigPath, contextName)
	restClientConfig, err := kubernetesConfig.ClientConfig()
	if err != nil {
		if clientcmd.IsEmptyConfig(err) {
			return nil, fmt.Errorf("couldn't find the kube config file, or file is empty (%s)\n"+
				"you can set alternative kube config file path by adding the kube-config-path field to the %s config file, err:  %w", kubeConfigPath, misc.Program, err)
		}
		if clientcmd.IsConfigurationInvalid(err) {
			return nil, fmt.Errorf("invalid kube config file (%s)\n"+
				"you can set alternative kube config file path by adding the kube-config-path field to the %s config file, err:  %w", kubeConfigPath, misc.Program, err)
		}

		return nil, fmt.Errorf("error while using kube config (%s)\n"+
			"you can set alternative kube config file path by adding the kube-config-path field to the %s config file, err:  %w", kubeConfigPath, misc.Program, err)
	}

	clientSet, err := getClientSet(restClientConfig)
	if err != nil {
		return nil, fmt.Errorf("error while using kube config (%s)\n"+
			"you can set alternative kube config file path by adding the kube-config-path field to the %s config file, err:  %w", kubeConfigPath, misc.Program, err)
	}

	log.Debug().
		Str("host", restClientConfig.Host).
		Str("api-path", restClientConfig.APIPath).
		Str("user-agent", restClientConfig.UserAgent).
		Msg("K8s client config.")

	return &Provider{
		clientSet:        clientSet,
		kubernetesConfig: kubernetesConfig,
		clientConfig:     *restClientConfig,
		managedBy:        misc.Program,
		createdBy:        misc.Program,
	}, nil
}

func (provider *Provider) DoesServiceExist(ctx context.Context, namespace string, name string) (bool, error) {
	serviceResource, err := provider.clientSet.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	return provider.doesResourceExist(serviceResource, err)
}

func (provider *Provider) doesResourceExist(resource interface{}, err error) (bool, error) {
	// Getting NotFound error is the expected behavior when a resource does not exist.
	if k8serrors.IsNotFound(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return resource != nil, nil
}

func (provider *Provider) listPodsImpl(ctx context.Context, regex *regexp.Regexp, namespaces []string, listOptions metav1.ListOptions) ([]core.Pod, error) {
	var pods []core.Pod
	for _, namespace := range namespaces {
		namespacePods, err := provider.clientSet.CoreV1().Pods(namespace).List(ctx, listOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to get pods in ns: [%s], %w", namespace, err)
		}

		pods = append(pods, namespacePods.Items...)
	}

	matchingPods := make([]core.Pod, 0)
	for _, pod := range pods {
		if regex.MatchString(pod.Name) {
			matchingPods = append(matchingPods, pod)
		}
	}
	return matchingPods, nil
}

func (provider *Provider) ListAllPodsMatchingRegex(ctx context.Context, regex *regexp.Regexp, namespaces []string) ([]core.Pod, error) {
	return provider.listPodsImpl(ctx, regex, namespaces, metav1.ListOptions{})
}

func (provider *Provider) ListAllRunningPodsMatchingRegex(ctx context.Context, regex *regexp.Regexp, namespaces []string) ([]core.Pod, error) {
	pods, err := provider.ListAllPodsMatchingRegex(ctx, regex, namespaces)
	if err != nil {
		return nil, err
	}

	matchingPods := make([]core.Pod, 0)
	for _, pod := range pods {
		if IsPodRunning(&pod) {
			matchingPods = append(matchingPods, pod)
		}
	}
	return matchingPods, nil
}

func (provider *Provider) ListPodsByAppLabel(ctx context.Context, namespaces string, labels map[string]string) ([]core.Pod, error) {
	pods, err := provider.clientSet.CoreV1().Pods(namespaces).List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(
			&metav1.LabelSelector{
				MatchLabels: labels,
			},
		),
	})
	if err != nil {
		return nil, err
	}

	return pods.Items, err
}

func (provider *Provider) GetPodLogs(ctx context.Context, namespace string, podName string, containerName string, grep string) (string, error) {
	podLogOpts := core.PodLogOptions{Container: containerName}
	req := provider.clientSet.CoreV1().Pods(namespace).GetLogs(podName, &podLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("error opening log stream on ns: %s, pod: %s, %w", namespace, podName, err)
	}
	defer podLogs.Close()
	buf := new(bytes.Buffer)
	if _, err = io.Copy(buf, podLogs); err != nil {
		return "", fmt.Errorf("error copy information from podLogs to buf, ns: %s, pod: %s, %w", namespace, podName, err)
	}

	if grep != "" {
		finder, err := reader.NewFinder(grep, true, true)
		if err != nil {
			panic(err)
		}

		read, err := reader.NewStdReader(bufio.NewReader(buf), []reader.Finder{finder})
		if err != nil {
			panic(err)
		}
		read.Run()
		result := read.Result()[0]

		log.Info().Str("namespace", namespace).Str("pod", podName).Str("container", containerName).Int("lines", len(result.Lines)).Str("grep", grep).Send()
		return strings.Join(result.MatchString, "\n"), nil
	} else {
		log.Info().Str("namespace", namespace).Str("pod", podName).Str("container", containerName).Send()
		return buf.String(), nil
	}
}

func (provider *Provider) GetNamespaceEvents(ctx context.Context, namespace string) (string, error) {
	eventList, err := provider.clientSet.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("error getting events on ns: %s, %w", namespace, err)
	}

	return eventList.String(), nil
}

// ValidateNotProxy We added this after a customer tried to run kubeshark from lens, which used len's kube config, which have cluster server configuration, which points to len's local proxy.
// The workaround was to use the user's local default kube config.
// For now - we are blocking the option to run kubeshark through a proxy to k8s server
func (provider *Provider) ValidateNotProxy() error {
	kubernetesUrl, err := url.Parse(provider.clientConfig.Host)
	if err != nil {
		log.Debug().Err(err).Msg("While parsing Kubernetes host!")
		return nil
	}

	restProxyClientConfig, _ := provider.kubernetesConfig.ClientConfig()
	restProxyClientConfig.Host = kubernetesUrl.Host

	clientProxySet, err := getClientSet(restProxyClientConfig)
	if err == nil {
		proxyServerVersion, err := clientProxySet.ServerVersion()
		if err != nil {
			return nil
		}

		if *proxyServerVersion == (version.Info{}) {
			return &ClusterBehindProxyError{}
		}
	}

	return nil
}

func (provider *Provider) GetKubernetesVersion() (*semver.SemVersion, error) {
	serverVersion, err := provider.clientSet.ServerVersion()
	if err != nil {
		log.Debug().Err(err).Msg("While getting Kubernetes server version!")
		return nil, err
	}

	serverVersionSemVer := semver.SemVersion(serverVersion.GitVersion)
	return &serverVersionSemVer, nil
}

func (provider *Provider) GetNamespaces() (namespaces []string) {
	if len(config.Config.Tap.Namespaces) > 0 {
		namespaces = utils.Unique(config.Config.Tap.Namespaces)
	} else {
		namespaceList, err := provider.clientSet.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Error().Err(err).Send()
			return
		}

		for _, ns := range namespaceList.Items {
			namespaces = append(namespaces, ns.Name)
		}
	}

	namespaces = utils.Diff(namespaces, config.Config.Tap.ExcludedNamespaces)

	return
}

func getClientSet(config *rest.Config) (*kubernetes.Clientset, error) {
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientSet, nil
}

func ValidateKubernetesVersion(serverVersionSemVer *semver.SemVersion) error {
	minKubernetesServerVersionSemVer := semver.SemVersion(MinKubernetesServerVersion)
	if minKubernetesServerVersionSemVer.GreaterThan(*serverVersionSemVer) {
		return fmt.Errorf("kubernetes server version %v is not supported, supporting only kubernetes server version of %v or higher", serverVersionSemVer, MinKubernetesServerVersion)
	}

	return nil
}

func loadKubernetesConfiguration(kubeConfigPath string, context string) clientcmd.ClientConfig {
	configPathList := filepath.SplitList(kubeConfigPath)
	configLoadingRules := &clientcmd.ClientConfigLoadingRules{}
	if len(configPathList) <= 1 {
		configLoadingRules.ExplicitPath = kubeConfigPath
	} else {
		configLoadingRules.Precedence = configPathList
	}
	contextName := context
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		configLoadingRules,
		&clientcmd.ConfigOverrides{
			CurrentContext: contextName,
		},
	)
}

func IsPodRunning(pod *core.Pod) bool {
	return pod.Status.Phase == core.PodRunning
}
