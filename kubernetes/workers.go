package kubernetes

import (
	"context"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/docker"
	"github.com/rs/zerolog/log"
	core "k8s.io/api/core/v1"
)

func CreateWorkers(
	kubernetesProvider *Provider,
	selfServiceAccountExists bool,
	ctx context.Context,
	namespace string,
	resources configStructs.ResourceRequirements,
	imagePullPolicy core.PullPolicy,
	imagePullSecrets []core.LocalObjectReference,
	serviceMesh bool,
	tls bool,
	debug bool,
) error {
	if config.Config.Tap.PersistentStorage {
		persistentVolumeClaim, err := kubernetesProvider.BuildPersistentVolumeClaim()
		if err != nil {
			return err
		}

		if _, err = kubernetesProvider.CreatePersistentVolumeClaim(
			ctx,
			namespace,
			persistentVolumeClaim,
		); err != nil {
			return err
		}
	}

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
