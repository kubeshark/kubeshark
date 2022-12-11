package resources

import (
	"context"

	"github.com/kubeshark/base/pkg/models"
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/docker"
	"github.com/kubeshark/kubeshark/errormessage"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/kubeshark"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	core "k8s.io/api/core/v1"
)

func CreateHubResources(ctx context.Context, kubernetesProvider *kubernetes.Provider, serializedKubesharkConfig string, isNsRestrictedMode bool, kubesharkResourcesNamespace string, maxEntriesDBSizeBytes int64, hubResources models.Resources, imagePullPolicy core.PullPolicy, logLevel zerolog.Level, profiler bool) (bool, error) {
	if !isNsRestrictedMode {
		if err := createKubesharkNamespace(ctx, kubernetesProvider, kubesharkResourcesNamespace); err != nil {
			return false, err
		}
	}

	if err := createKubesharkConfigmap(ctx, kubernetesProvider, serializedKubesharkConfig, kubesharkResourcesNamespace); err != nil {
		return false, err
	}

	kubesharkServiceAccountExists, err := createRBACIfNecessary(ctx, kubernetesProvider, isNsRestrictedMode, kubesharkResourcesNamespace, []string{"pods", "services", "endpoints"})
	if err != nil {
		log.Warn().Err(errormessage.FormatError(err)).Msg("Failed to ensure the resources required for IP resolving. Kubeshark will not resolve target IPs to names.")
	}

	var serviceAccountName string
	if kubesharkServiceAccountExists {
		serviceAccountName = kubernetes.ServiceAccountName
	} else {
		serviceAccountName = ""
	}

	opts := &kubernetes.HubOptions{
		Namespace:             kubesharkResourcesNamespace,
		PodName:               kubernetes.HubPodName,
		PodImage:              docker.GetHubImage(),
		KratosImage:           "",
		KetoImage:             "",
		ServiceAccountName:    serviceAccountName,
		IsNamespaceRestricted: isNsRestrictedMode,
		MaxEntriesDBSizeBytes: maxEntriesDBSizeBytes,
		Resources:             hubResources,
		ImagePullPolicy:       imagePullPolicy,
		LogLevel:              logLevel,
		Profiler:              profiler,
	}

	frontOpts := &kubernetes.HubOptions{
		Namespace:             kubesharkResourcesNamespace,
		PodName:               kubernetes.FrontPodName,
		PodImage:              docker.GetWorkerImage(),
		KratosImage:           "",
		KetoImage:             "",
		ServiceAccountName:    serviceAccountName,
		IsNamespaceRestricted: isNsRestrictedMode,
		MaxEntriesDBSizeBytes: maxEntriesDBSizeBytes,
		Resources:             hubResources,
		ImagePullPolicy:       imagePullPolicy,
		LogLevel:              logLevel,
		Profiler:              profiler,
	}

	if err := createKubesharkHubPod(ctx, kubernetesProvider, opts); err != nil {
		return kubesharkServiceAccountExists, err
	}

	if err := createFrontPod(ctx, kubernetesProvider, frontOpts); err != nil {
		return kubesharkServiceAccountExists, err
	}

	_, err = kubernetesProvider.CreateService(ctx, kubesharkResourcesNamespace, kubernetes.HubServiceName, kubernetes.HubServiceName, 80, int32(config.Config.Hub.PortForward.DstPort), int32(config.Config.Hub.PortForward.SrcPort))
	if err != nil {
		return kubesharkServiceAccountExists, err
	}

	log.Info().Str("service", kubernetes.HubServiceName).Msg("Successfully created a service.")

	_, err = kubernetesProvider.CreateService(ctx, kubesharkResourcesNamespace, kubernetes.FrontServiceName, kubernetes.FrontServiceName, 80, int32(config.Config.Front.PortForward.DstPort), int32(config.Config.Front.PortForward.SrcPort))
	if err != nil {
		return kubesharkServiceAccountExists, err
	}

	log.Info().Str("service", kubernetes.FrontServiceName).Msg("Successfully created a service.")

	return kubesharkServiceAccountExists, nil
}

func createKubesharkNamespace(ctx context.Context, kubernetesProvider *kubernetes.Provider, kubesharkResourcesNamespace string) error {
	_, err := kubernetesProvider.CreateNamespace(ctx, kubesharkResourcesNamespace)
	return err
}

func createKubesharkConfigmap(ctx context.Context, kubernetesProvider *kubernetes.Provider, serializedKubesharkConfig string, kubesharkResourcesNamespace string) error {
	err := kubernetesProvider.CreateConfigMap(ctx, kubesharkResourcesNamespace, kubernetes.ConfigMapName, serializedKubesharkConfig)
	return err
}

func createRBACIfNecessary(ctx context.Context, kubernetesProvider *kubernetes.Provider, isNsRestrictedMode bool, kubesharkResourcesNamespace string, resources []string) (bool, error) {
	if !isNsRestrictedMode {
		if err := kubernetesProvider.CreateKubesharkRBAC(ctx, kubesharkResourcesNamespace, kubernetes.ServiceAccountName, kubernetes.ClusterRoleName, kubernetes.ClusterRoleBindingName, kubeshark.RBACVersion, resources); err != nil {
			return false, err
		}
	} else {
		if err := kubernetesProvider.CreateKubesharkRBACNamespaceRestricted(ctx, kubesharkResourcesNamespace, kubernetes.ServiceAccountName, kubernetes.RoleName, kubernetes.RoleBindingName, kubeshark.RBACVersion); err != nil {
			return false, err
		}
	}

	return true, nil
}

func createKubesharkHubPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, opts *kubernetes.HubOptions) error {
	pod, err := kubernetesProvider.BuildHubPod(opts, false, "", false)
	if err != nil {
		return err
	}
	if _, err = kubernetesProvider.CreatePod(ctx, opts.Namespace, pod); err != nil {
		return err
	}
	log.Info().Str("pod", pod.Name).Msg("Successfully created a pod.")
	return nil
}

func createFrontPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, opts *kubernetes.HubOptions) error {
	pod, err := kubernetesProvider.BuildFrontPod(opts, false, "", false)
	if err != nil {
		return err
	}
	if _, err = kubernetesProvider.CreatePod(ctx, opts.Namespace, pod); err != nil {
		return err
	}
	log.Info().Str("pod", pod.Name).Msg("Successfully created a pod.")
	return nil
}
