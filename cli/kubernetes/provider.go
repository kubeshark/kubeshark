package kubernetes

import (
	_ "bytes"
	"context"
	"fmt"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	_ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	_ "k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

type Provider struct {
	clientSet        *kubernetes.Clientset
	kubernetesConfig clientcmd.ClientConfig
	clientConfig     restclient.Config
	Namespace        string
}

func NewProvider(kubeConfigPath string, overrideNamespace string) *Provider {
	kubernetesConfig := loadKubernetesConfiguration(kubeConfigPath)
	restClientConfig, err := kubernetesConfig.ClientConfig()
	if err != nil {
		panic(err.Error())
	}
	clientSet := getClientSet(restClientConfig)

	var namespace string
	if len(overrideNamespace) > 0 {
		namespace = overrideNamespace
	} else {
		configuredNamespace, _, err := kubernetesConfig.Namespace()
		if err != nil {
			panic(err)
		}
		namespace = configuredNamespace
	}

	return &Provider{
		clientSet:        clientSet,
		kubernetesConfig: kubernetesConfig,
		clientConfig:     *restClientConfig,
		Namespace:        namespace,
	}
}

func (provider *Provider) GetPodWatcher(ctx context.Context) watch.Interface {
	watcher, err := provider.clientSet.CoreV1().Pods(provider.Namespace).Watch(ctx, metav1.ListOptions{Watch: true})
	if err != nil {
		panic(err.Error())
	}
	return watcher
}

func (provider *Provider) GetPods(ctx context.Context) {
	pods, err := provider.clientSet.CoreV1().Pods(provider.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in Namespace %s\n", len(pods.Items), provider.Namespace)
}

func (provider *Provider) CreatePod(ctx context.Context, podName string, podImage string) (*core.Pod, error) {
	privileged := true
	pod := &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: provider.Namespace,
		},
		Spec: core.PodSpec{
			HostNetwork: true,  // very important to make passive tapper see traffic
			Containers: []core.Container{
				{
					Name:            podName,
					Image:           podImage,
					ImagePullPolicy: core.PullAlways,
					SecurityContext: &core.SecurityContext{
						Privileged: &privileged, // must be privileged to get node level traffic
					},
				},
			},
			TerminationGracePeriodSeconds: new(int64),
		},
	}
	return provider.clientSet.CoreV1().Pods(provider.Namespace).Create(ctx, pod, metav1.CreateOptions{})
}

func (provider *Provider) RemovePod(ctx context.Context, podName string) {
	provider.clientSet.CoreV1().Pods(provider.Namespace).Delete(ctx, podName, metav1.DeleteOptions{})
}

func getClientSet(config *restclient.Config) *kubernetes.Clientset {
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientSet
}

func loadKubernetesConfiguration(kubeConfigPath string) clientcmd.ClientConfig {
	if kubeConfigPath == "" {
		home := homedir.HomeDir()
		kubeConfigPath = filepath.Join(home, ".kube", "config")
	}

	configPathList := filepath.SplitList(kubeConfigPath)
	configLoadingRules := &clientcmd.ClientConfigLoadingRules{}
	if len(configPathList) <= 1 {
		configLoadingRules.ExplicitPath = kubeConfigPath
	} else {
		configLoadingRules.Precedence = configPathList
	}
	contextName := ""
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		configLoadingRules,
		&clientcmd.ConfigOverrides{
			CurrentContext: contextName,
		},
	)
}
