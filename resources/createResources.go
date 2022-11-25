package resources

import (
	"context"
	"fmt"
	"log"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/errormessage"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/kubeshark"
	"github.com/kubeshark/kubeshark/uiUtils"
	"github.com/kubeshark/worker/models"
	"github.com/op/go-logging"
	core "k8s.io/api/core/v1"
)

func CreateTapKubesharkResources(ctx context.Context, kubernetesProvider *kubernetes.Provider, serializedKubesharkConfig string, isNsRestrictedMode bool, kubesharkResourcesNamespace string, agentImage string, maxEntriesDBSizeBytes int64, apiServerResources models.Resources, imagePullPolicy core.PullPolicy, logLevel logging.Level, profiler bool) (bool, error) {
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
		log.Printf(uiUtils.Warning, fmt.Sprintf("Failed to ensure the resources required for IP resolving. Kubeshark will not resolve target IPs to names. error: %v", errormessage.FormatError(err)))
	}

	var serviceAccountName string
	if kubesharkServiceAccountExists {
		serviceAccountName = kubernetes.ServiceAccountName
	} else {
		serviceAccountName = ""
	}

	opts := &kubernetes.ApiServerOptions{
		Namespace:             kubesharkResourcesNamespace,
		PodName:               kubernetes.ApiServerPodName,
		PodImage:              agentImage,
		KratosImage:           "",
		KetoImage:             "",
		ServiceAccountName:    serviceAccountName,
		IsNamespaceRestricted: isNsRestrictedMode,
		MaxEntriesDBSizeBytes: maxEntriesDBSizeBytes,
		Resources:             apiServerResources,
		ImagePullPolicy:       imagePullPolicy,
		LogLevel:              logLevel,
		Profiler:              profiler,
	}

	if err := createKubesharkApiServerPod(ctx, kubernetesProvider, opts); err != nil {
		return kubesharkServiceAccountExists, err
	}

	if err := createFrontPod(ctx, kubernetesProvider, opts); err != nil {
		return kubesharkServiceAccountExists, err
	}

	_, err = kubernetesProvider.CreateService(ctx, kubesharkResourcesNamespace, kubernetes.ApiServerPodName, kubernetes.ApiServerPodName, 80, int32(config.Config.Hub.PortForward.DstPort), int32(config.Config.Hub.PortForward.SrcPort))
	if err != nil {
		return kubesharkServiceAccountExists, err
	}

	_, err = kubernetesProvider.CreateService(ctx, kubesharkResourcesNamespace, "front", "front", 80, int32(config.Config.Front.PortForward.DstPort), int32(config.Config.Front.PortForward.SrcPort))
	if err != nil {
		return kubesharkServiceAccountExists, err
	}

	log.Printf("Successfully created service: %s", kubernetes.ApiServerPodName)

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

func createKubesharkApiServerPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, opts *kubernetes.ApiServerOptions) error {
	pod, err := kubernetesProvider.BuildApiServerPod(opts, false, "", false)
	if err != nil {
		return err
	}
	if _, err = kubernetesProvider.CreatePod(ctx, opts.Namespace, pod); err != nil {
		return err
	}
	log.Printf("Successfully created pod: [%s]", pod.Name)
	return nil
}

func createFrontPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, opts *kubernetes.ApiServerOptions) error {
	pod, err := kubernetesProvider.BuildFrontPod(opts, false, "", false)
	if err != nil {
		return err
	}
	if _, err = kubernetesProvider.CreatePod(ctx, opts.Namespace, pod); err != nil {
		return err
	}
	log.Printf("Successfully created pod: [%s]", pod.Name)
	return nil
}
