package kubernetes

import (
	"sync"

	"github.com/up9inc/mizu/shared/kubernetes"
)

var lock = &sync.Mutex{}

var kubernetesProvider *kubernetes.Provider

func GetKubernetesProvider() (*kubernetes.Provider, error) {
	if kubernetesProvider == nil {
		lock.Lock()
		defer lock.Unlock()

		if kubernetesProvider == nil {
			var err error
			kubernetesProvider, err = kubernetes.NewProviderInCluster()
			if err != nil {
				return nil, err
			}
		}
	}

	return kubernetesProvider, nil
}
