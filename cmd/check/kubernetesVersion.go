package check

import (
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/semver"
	"github.com/rs/zerolog/log"
)

func KubernetesVersion(kubernetesVersion *semver.SemVersion) bool {
	log.Info().Msg("[kubernetes-api]")

	if err := kubernetes.ValidateKubernetesVersion(kubernetesVersion); err != nil {
		log.Error().Err(err).Msg("Not running the minimum Kubernetes API version!")
		return false
	}

	log.Info().Msg("Running the minimum Kubernetes API version")
	return true
}
