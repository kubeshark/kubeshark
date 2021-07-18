package resolver

import (
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	_ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"os"
)

func NewFromInCluster(errOut chan error) (*Resolver, error) {
	config, err := restclient.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Resolver{clientConfig: config, clientSet: clientset, nameMap: make(map[string]string), serviceMap: make(map[string]string), errOut: errOut}, nil
}

func NewFromOutOfCluster(kubeConfigPath string, errOut chan error) (*Resolver, error) {
	if kubeConfigPath == "" {
		env := os.Getenv("KUBECONFIG") 
		if env != "" {
			kubeConfigPath = env
		} else {
			home := homedir.HomeDir()
			kubeConfigPath = filepath.Join(home, ".kube", "config")
		}
	}

	configPathList := filepath.SplitList(kubeConfigPath)
	configLoadingRules := &clientcmd.ClientConfigLoadingRules{}
	if len(configPathList) <= 1 {
		configLoadingRules.ExplicitPath = kubeConfigPath
	} else {
		configLoadingRules.Precedence = configPathList
	}
	contextName := ""
	clientConfigLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		configLoadingRules,
		&clientcmd.ConfigOverrides{
			CurrentContext: contextName,
		},
	)
	clientConfig, err := clientConfigLoader.ClientConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return &Resolver{clientConfig: clientConfig, clientSet: clientset, nameMap: make(map[string]string), serviceMap: make(map[string]string), errOut: errOut}, nil
}

func NewFromExisting(clientConfig *restclient.Config, clientSet *kubernetes.Clientset, errOut chan error) *Resolver {
	return &Resolver{clientConfig: clientConfig, clientSet: clientSet, nameMap: make(map[string]string), serviceMap: make(map[string]string), errOut: errOut}
}
