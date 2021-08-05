package resolver

import (
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	_ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
	restclient "k8s.io/client-go/rest"
)

func NewFromInCluster(namespace string, errOut chan error) (*Resolver, error) {
	config, err := restclient.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Resolver{
		clientConfig: config,
		clientSet: clientset,
		nameMap: make(map[string]string),
		serviceMap: make(map[string]string),
		errOut: errOut,
		namespace: namespace,
	}, nil
}
