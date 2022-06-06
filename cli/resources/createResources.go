package resources

import (
	"context"
	"fmt"

	"github.com/op/go-logging"
	"github.com/up9inc/mizu/cli/errormessage"
	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/kubernetes"
	core "k8s.io/api/core/v1"
)

func CreateTapMizuResources(ctx context.Context, kubernetesProvider *kubernetes.Provider, serializedValidationRules string, serializedContract string, serializedMizuConfig string, isNsRestrictedMode bool, mizuResourcesNamespace string, agentImage string, maxEntriesDBSizeBytes int64, apiServerResources shared.Resources, imagePullPolicy core.PullPolicy, logLevel logging.Level, profiler bool) (bool, error) {
	if !isNsRestrictedMode {
		if err := createMizuNamespace(ctx, kubernetesProvider, mizuResourcesNamespace); err != nil {
			return false, err
		}
	}

	if err := createMizuConfigmap(ctx, kubernetesProvider, serializedValidationRules, serializedContract, serializedMizuConfig, mizuResourcesNamespace); err != nil {
		return false, err
	}

	mizuServiceAccountExists, err := createRBACIfNecessary(ctx, kubernetesProvider, isNsRestrictedMode, mizuResourcesNamespace, []string{"pods", "services", "endpoints"})
	if err != nil {
		logger.Log.Warningf(uiUtils.Warning, fmt.Sprintf("Failed to ensure the resources required for IP resolving. Mizu will not resolve target IPs to names. error: %v", errormessage.FormatError(err)))
	}

	var serviceAccountName string
	if mizuServiceAccountExists {
		serviceAccountName = kubernetes.ServiceAccountName
	} else {
		serviceAccountName = ""
	}

	opts := &kubernetes.ApiServerOptions{
		Namespace:             mizuResourcesNamespace,
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

	if err := createMizuApiServerPod(ctx, kubernetesProvider, opts); err != nil {
		return mizuServiceAccountExists, err
	}

	_, err = kubernetesProvider.CreateService(ctx, mizuResourcesNamespace, kubernetes.ApiServerPodName, kubernetes.ApiServerPodName)
	if err != nil {
		return mizuServiceAccountExists, err
	}

	logger.Log.Debugf("Successfully created service: %s", kubernetes.ApiServerPodName)

	return mizuServiceAccountExists, nil
}

func createMizuNamespace(ctx context.Context, kubernetesProvider *kubernetes.Provider, mizuResourcesNamespace string) error {
	_, err := kubernetesProvider.CreateNamespace(ctx, mizuResourcesNamespace)
	return err
}

func createMizuConfigmap(ctx context.Context, kubernetesProvider *kubernetes.Provider, serializedValidationRules string, serializedContract string, serializedMizuConfig string, mizuResourcesNamespace string) error {
	err := kubernetesProvider.CreateConfigMap(ctx, mizuResourcesNamespace, kubernetes.ConfigMapName, serializedValidationRules, serializedContract, serializedMizuConfig)
	return err
}

func createRBACIfNecessary(ctx context.Context, kubernetesProvider *kubernetes.Provider, isNsRestrictedMode bool, mizuResourcesNamespace string, resources []string) (bool, error) {
	if !isNsRestrictedMode {
		if err := kubernetesProvider.CreateMizuRBAC(ctx, mizuResourcesNamespace, kubernetes.ServiceAccountName, kubernetes.ClusterRoleName, kubernetes.ClusterRoleBindingName, mizu.RBACVersion, resources); err != nil {
			return false, err
		}
	} else {
		if err := kubernetesProvider.CreateMizuRBACNamespaceRestricted(ctx, mizuResourcesNamespace, kubernetes.ServiceAccountName, kubernetes.RoleName, kubernetes.RoleBindingName, mizu.RBACVersion); err != nil {
			return false, err
		}
	}

	return true, nil
}

func createMizuApiServerPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, opts *kubernetes.ApiServerOptions) error {
	pod, err := kubernetesProvider.GetMizuApiServerPodObject(opts, false, "", false)
	if err != nil {
		return err
	}
	if _, err = kubernetesProvider.CreatePod(ctx, opts.Namespace, pod); err != nil {
		return err
	}
	logger.Log.Debugf("Successfully created API server pod: %s", kubernetes.ApiServerPodName)
	return nil
}
