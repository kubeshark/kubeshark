package kubernetes

import (
	"context"

	"github.com/kubeshark/kubeshark/docker"
	"github.com/rs/zerolog/log"
	core "k8s.io/api/core/v1"
)

func CreateWorkers(
	kubernetesProvider *Provider,
	selfServiceAccountExists bool,
	ctx context.Context,
	namespace string,
	resources Resources,
	imagePullPolicy core.PullPolicy,
	imagePullSecrets []core.LocalObjectReference,
	serviceMesh bool,
	tls bool,
	debug bool,
) error {
	image := docker.GetWorkerImage()

	var serviceAccountName string
	if selfServiceAccountExists {
		serviceAccountName = ServiceAccountName
	} else {
		serviceAccountName = ""
	}

	log.Info().Msg("Creating the worker DaemonSet...")

	if err := kubernetesProvider.ApplyWorkerDaemonSet(
		ctx,
		namespace,
		WorkerDaemonSetName,
		image,
		WorkerPodName,
		serviceAccountName,
		resources,
		imagePullPolicy,
		imagePullSecrets,
		serviceMesh,
		tls,
		debug,
	); err != nil {
		return err
	}

	log.Info().Msg("Successfully created the worker DaemonSet.")

	return nil
}
