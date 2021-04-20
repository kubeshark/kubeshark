package kubernetes

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

type kubernetesProvider struct {
	clientSet *kubernetes.Clientset
	kubernetesConfig clientcmd.ClientConfig
}

func New(kubeConfigPath string) *kubernetesProvider {
	kubernetesConfig := loadKubernetesConfiguration(kubeConfigPath)
	restClientConfig, err := kubernetesConfig.ClientConfig()
	if err != nil {
		panic(err.Error())
	}
	clientSet := getClientAndConfig(restClientConfig)
	return &kubernetesProvider{clientSet: clientSet, kubernetesConfig: kubernetesConfig}
}

func (provider *kubernetesProvider) GetPodWatcher(ctx context.Context) watch.Interface {
	watcher, err := provider.clientSet.CoreV1().Pods(provider.getCurrentNamespace()).Watch(ctx, metav1.ListOptions{Watch: true})
	if err != nil {
		panic(err.Error())
	}
	return watcher
}

func (provider *kubernetesProvider) GetPods() {
	namespace := provider.getCurrentNamespace()
	pods, err := provider.clientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in namespace %s\n", len(pods.Items), namespace)
}

func (provider *kubernetesProvider) getCurrentNamespace() string {
	namespace, _, err := provider.kubernetesConfig.Namespace()
	if err != nil {
		panic(err.Error())
	}
	return namespace
}

func getClientAndConfig(config *restclient.Config) *kubernetes.Clientset {
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
