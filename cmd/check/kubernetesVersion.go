package check

import (
	"fmt"

	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/semver"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/rs/zerolog/log"
)

func KubernetesVersion(kubernetesVersion *semver.SemVersion) bool {
	log.Info().Str("procedure", "kubernetes-version").Msg("Checking:")

	if err := kubernetes.ValidateKubernetesVersion(kubernetesVersion); err != nil {
		log.Error().Str("k8s-version", string(*kubernetesVersion)).Err(err).Msg(fmt.Sprintf(utils.Red, "The cluster does not have the minimum required Kubernetes API version!"))
		return false
	}

	log.Info().Str("k8s-version", string(*kubernetesVersion)).Msg("Minimum required Kubernetes API version is passed.")
	return true
}
