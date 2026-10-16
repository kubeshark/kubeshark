package helm

import (
	"path/filepath"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// restClientGetter gives Helm a Kubernetes client built from a kubeconfig path
// list, such as the value of $KUBECONFIG. Helm's own config flags treat the whole
// value as a single file name, so a list of paths fails with "no such file".
type restClientGetter struct {
	loader clientcmd.ClientConfig
}

func newRESTClientGetter(kubeConfigPath string, namespace string) *restClientGetter {
	loadingRules := &clientcmd.ClientConfigLoadingRules{}
	if paths := filepath.SplitList(kubeConfigPath); len(paths) <= 1 {
		loadingRules.ExplicitPath = kubeConfigPath
	} else {
		loadingRules.Precedence = paths
	}

	overrides := &clientcmd.ConfigOverrides{
		Context: clientcmdapi.Context{Namespace: namespace},
	}

	return &restClientGetter{
		loader: clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides),
	}
}

func (g *restClientGetter) ToRESTConfig() (*rest.Config, error) {
	return g.loader.ClientConfig()
}

func (g *restClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	config, err := g.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	client, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	return memory.NewMemCacheClient(client), nil
}

func (g *restClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	client, err := g.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	return restmapper.NewDeferredDiscoveryRESTMapper(client), nil
}

func (g *restClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return g.loader
}
