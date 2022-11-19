package resources

import (
	"context"
	"fmt"

	"github.com/kubeshark/kubeshark/cli/errormessage"
	"github.com/kubeshark/kubeshark/cli/kubeshark"
	"github.com/kubeshark/kubeshark/cli/uiUtils"
	"github.com/kubeshark/kubeshark/logger"
	"github.com/kubeshark/kubeshark/shared"
	"github.com/kubeshark/kubeshark/shared/kubernetes"
	"github.com/op/go-logging"
	core "k8s.io/api/core/v1"
)

func CreateTapKubesharkResources(ctx context.Context, kubernetesProvider *kubernetes.Provider, serializedKubesharkConfig string, isNsRestrictedMode bool, kubesharkResourcesNamespace string, agentImage string, maxEntriesDBSizeBytes int64, apiServerResources shared.Resources, imagePullPolicy core.PullPolicy, logLevel logging.Level, profiler bool) (bool, error) {
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
		logger.Log.Warningf(uiUtils.Warning, fmt.Sprintf("Failed to ensure the resources required for IP resolving. Kubeshark will not resolve target IPs to names. error: %v", errormessage.FormatError(err)))
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

	_, err = kubernetesProvider.CreateService(ctx, kubesharkResourcesNamespace, kubernetes.ApiServerPodName, kubernetes.ApiServerPodName)
	if err != nil {
		return kubesharkServiceAccountExists, err
	}

	logger.Log.Debugf("Successfully created service: %s", kubernetes.ApiServerPodName)

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
	pod, err := kubernetesProvider.GetKubesharkApiServerPodObject(opts, false, "", false)
	if err != nil {
		return err
	}
	if _, err = kubernetesProvider.CreatePod(ctx, opts.Namespace, pod); err != nil {
		return err
	}
	logger.Log.Debugf("Successfully created API server pod: %s", kubernetes.ApiServerPodName)
	return nil
}
