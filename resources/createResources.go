package resources

import (
	"context"
	"fmt"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/docker"
	"github.com/kubeshark/kubeshark/errormessage"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/misc"
	"github.com/rs/zerolog/log"
	core "k8s.io/api/core/v1"
)

func CreateHubResources(ctx context.Context, kubernetesProvider *kubernetes.Provider, isNsRestrictedMode bool, selfNamespace string, hubResources configStructs.ResourceRequirements, imagePullPolicy core.PullPolicy, imagePullSecrets []core.LocalObjectReference, debug bool) (bool, error) {
	if !isNsRestrictedMode {
		if err := createSelfNamespace(ctx, kubernetesProvider, selfNamespace); err != nil {
			log.Debug().Err(err).Send()
		}
	}

	err := kubernetesProvider.CreateSelfRBAC(ctx, selfNamespace)
	var selfServiceAccountExists bool
	if err != nil {
		selfServiceAccountExists = true
		log.Warn().Err(errormessage.FormatError(err)).Msg(fmt.Sprintf("Failed to ensure the resources required for IP resolving. %s will not resolve target IPs to names.", misc.Software))
	}

	hubOpts := &kubernetes.PodOptions{
		Namespace:          selfNamespace,
		PodName:            kubernetes.HubPodName,
		PodImage:           docker.GetHubImage(),
		ServiceAccountName: kubernetes.ServiceAccountName,
		Resources:          hubResources,
		ImagePullPolicy:    imagePullPolicy,
		ImagePullSecrets:   imagePullSecrets,
		Debug:              debug,
	}

	frontOpts := &kubernetes.PodOptions{
		Namespace:          selfNamespace,
		PodName:            kubernetes.FrontPodName,
		PodImage:           docker.GetWorkerImage(),
		ServiceAccountName: kubernetes.ServiceAccountName,
		Resources:          hubResources,
		ImagePullPolicy:    imagePullPolicy,
		ImagePullSecrets:   imagePullSecrets,
		Debug:              debug,
	}

	if err := createSelfHubPod(ctx, kubernetesProvider, hubOpts); err != nil {
		return selfServiceAccountExists, err
	}

	if err := createFrontPod(ctx, kubernetesProvider, frontOpts); err != nil {
		return selfServiceAccountExists, err
	}

	_, err = kubernetesProvider.CreateService(ctx, selfNamespace, kubernetesProvider.BuildHubService(selfNamespace))
	if err != nil {
		return selfServiceAccountExists, err
	}
	log.Info().Str("service", kubernetes.HubServiceName).Msg("Successfully created a service.")

	_, err = kubernetesProvider.CreateService(ctx, selfNamespace, kubernetesProvider.BuildFrontService(selfNamespace))
	if err != nil {
		return selfServiceAccountExists, err
	}
	log.Info().Str("service", kubernetes.FrontServiceName).Msg("Successfully created a service.")

	_, err = kubernetesProvider.CreateIngressClass(ctx, kubernetesProvider.BuildIngressClass())
	if err != nil {
		return selfServiceAccountExists, err
	}
	log.Info().Str("ingress-class", kubernetes.IngressClassName).Msg("Successfully created an ingress class.")

	_, err = kubernetesProvider.CreateIngress(ctx, selfNamespace, kubernetesProvider.BuildIngress())
	if err != nil {
		return selfServiceAccountExists, err
	}
	log.Info().Str("ingress", kubernetes.IngressName).Msg("Successfully created an ingress.")

	return selfServiceAccountExists, nil
}

func createSelfNamespace(ctx context.Context, kubernetesProvider *kubernetes.Provider, selfNamespace string) error {
	_, err := kubernetesProvider.CreateNamespace(ctx, kubernetesProvider.BuildNamespace(selfNamespace))
	return err
}

func createSelfHubPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, opts *kubernetes.PodOptions) error {
	pod, err := kubernetesProvider.BuildHubPod(opts)
	if err != nil {
		return err
	}
	if _, err = kubernetesProvider.CreatePod(ctx, opts.Namespace, pod); err != nil {
		return err
	}
	log.Info().Str("pod", pod.Name).Msg("Successfully created a pod.")
	return nil
}

func createFrontPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, opts *kubernetes.PodOptions) error {
	pod, err := kubernetesProvider.BuildFrontPod(opts, config.Config.Tap.Proxy.Host, fmt.Sprintf("%d", config.Config.Tap.Proxy.Hub.Port))
	if err != nil {
		return err
	}
	if _, err = kubernetesProvider.CreatePod(ctx, opts.Namespace, pod); err != nil {
		return err
	}
	log.Info().Str("pod", pod.Name).Msg("Successfully created a pod.")
	return nil
}
