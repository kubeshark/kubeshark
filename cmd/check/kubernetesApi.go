package check

import (
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/semver"
	"github.com/rs/zerolog/log"
)

func KubernetesApi() (*kubernetes.Provider, *semver.SemVersion, bool) {
	log.Info().Str("procedure", "kubernetes-api").Msg("Checking:")

	kubernetesProvider, err := kubernetes.NewProvider(config.Config.KubeConfigPath(), config.Config.Kube.Context)
	if err != nil {
		log.Error().Err(err).Msg("Can't initialize the client!")
		return nil, nil, false
	}
	log.Info().Msg("Initialization of the client is passed.")

	kubernetesVersion, err := kubernetesProvider.GetKubernetesVersion()
	if err != nil {
		log.Error().Err(err).Msg("Can't query the Kubernetes API!")
		return nil, nil, false
	}
	log.Info().Msg("Querying the Kubernetes API is passed.")

	return kubernetesProvider, kubernetesVersion, true
}
