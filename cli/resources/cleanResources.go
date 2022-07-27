package resources

import (
	"context"
	"fmt"

	"github.com/up9inc/mizu/cli/errormessage"
	"github.com/up9inc/mizu/cli/mizu/fsUtils"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/cli/utils"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/shared/kubernetes"
	"k8s.io/apimachinery/pkg/util/wait"
)

func CleanUpMizuResources(ctx context.Context, cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider, isNsRestrictedMode bool, mizuResourcesNamespace string) {
	logger.Log.Infof("\nRemoving mizu resources")

	var leftoverResources []string

	if isNsRestrictedMode {
		leftoverResources = cleanUpRestrictedMode(ctx, kubernetesProvider, mizuResourcesNamespace)
	} else {
		leftoverResources = cleanUpNonRestrictedMode(ctx, cancel, kubernetesProvider, mizuResourcesNamespace)
	}

	if len(leftoverResources) > 0 {
		errMsg := fmt.Sprintf("Failed to remove the following resources, for more info check logs at %s:", fsUtils.GetLogFilePath())
		for _, resource := range leftoverResources {
			errMsg += "\n- " + resource
		}
		logger.Log.Errorf(uiUtils.Error, errMsg)
	}
}

func cleanUpNonRestrictedMode(ctx context.Context, cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider, mizuResourcesNamespace string) []string {
	leftoverResources := make([]string, 0)

	if err := kubernetesProvider.RemoveNamespace(ctx, mizuResourcesNamespace); err != nil {
		resourceDesc := fmt.Sprintf("Namespace %s", mizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	} else {
		defer waitUntilNamespaceDeleted(ctx, cancel, kubernetesProvider, mizuResourcesNamespace)
	}

	if resources, err := kubernetesProvider.ListManagedClusterRoles(ctx); err != nil {
		resourceDesc := "ClusterRoles"
		handleDeletionError(err, resourceDesc, &leftoverResources)
	} else {
		for _, resource := range resources.Items {
			if err := kubernetesProvider.RemoveClusterRole(ctx, resource.Name); err != nil {
				resourceDesc := fmt.Sprintf("ClusterRole %s", resource.Name)
				handleDeletionError(err, resourceDesc, &leftoverResources)
			}
		}
	}

	if resources, err := kubernetesProvider.ListManagedClusterRoleBindings(ctx); err != nil {
		resourceDesc := "ClusterRoleBindings"
		handleDeletionError(err, resourceDesc, &leftoverResources)
	} else {
		for _, resource := range resources.Items {
			if err := kubernetesProvider.RemoveClusterRoleBinding(ctx, resource.Name); err != nil {
				resourceDesc := fmt.Sprintf("ClusterRoleBinding %s", resource.Name)
				handleDeletionError(err, resourceDesc, &leftoverResources)
			}
		}
	}

	return leftoverResources
}

func waitUntilNamespaceDeleted(ctx context.Context, cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider, mizuResourcesNamespace string) {
	// Call cancel if a terminating signal was received. Allows user to skip the wait.
	go func() {
		utils.WaitForFinish(ctx, cancel)
	}()

	if err := kubernetesProvider.WaitUtilNamespaceDeleted(ctx, mizuResourcesNamespace); err != nil {
		switch {
		case ctx.Err() == context.Canceled:
			logger.Log.Debugf("Do nothing. User interrupted the wait")
		case err == wait.ErrWaitTimeout:
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Timeout while removing Namespace %s", mizuResourcesNamespace))
		default:
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error while waiting for Namespace %s to be deleted: %v", mizuResourcesNamespace, errormessage.FormatError(err)))
		}
	}
}

func cleanUpRestrictedMode(ctx context.Context, kubernetesProvider *kubernetes.Provider, mizuResourcesNamespace string) []string {
	leftoverResources := make([]string, 0)

	if err := kubernetesProvider.RemoveService(ctx, mizuResourcesNamespace, kubernetes.ApiServerPodName); err != nil {
		resourceDesc := fmt.Sprintf("Service %s in namespace %s", kubernetes.ApiServerPodName, mizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveDaemonSet(ctx, mizuResourcesNamespace, kubernetes.TapperDaemonSetName); err != nil {
		resourceDesc := fmt.Sprintf("DaemonSet %s in namespace %s", kubernetes.TapperDaemonSetName, mizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveConfigMap(ctx, mizuResourcesNamespace, kubernetes.ConfigMapName); err != nil {
		resourceDesc := fmt.Sprintf("ConfigMap %s in namespace %s", kubernetes.ConfigMapName, mizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if resources, err := kubernetesProvider.ListManagedServiceAccounts(ctx, mizuResourcesNamespace); err != nil {
		resourceDesc := fmt.Sprintf("ServiceAccounts in namespace %s", mizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	} else {
		for _, resource := range resources.Items {
			if err := kubernetesProvider.RemoveServiceAccount(ctx, mizuResourcesNamespace, resource.Name); err != nil {
				resourceDesc := fmt.Sprintf("ServiceAccount %s in namespace %s", resource.Name, mizuResourcesNamespace)
				handleDeletionError(err, resourceDesc, &leftoverResources)
			}
		}
	}

	if resources, err := kubernetesProvider.ListManagedRoles(ctx, mizuResourcesNamespace); err != nil {
		resourceDesc := fmt.Sprintf("Roles in namespace %s", mizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	} else {
		for _, resource := range resources.Items {
			if err := kubernetesProvider.RemoveRole(ctx, mizuResourcesNamespace, resource.Name); err != nil {
				resourceDesc := fmt.Sprintf("Role %s in namespace %s", resource.Name, mizuResourcesNamespace)
				handleDeletionError(err, resourceDesc, &leftoverResources)
			}
		}
	}

	if resources, err := kubernetesProvider.ListManagedRoleBindings(ctx, mizuResourcesNamespace); err != nil {
		resourceDesc := fmt.Sprintf("RoleBindings in namespace %s", mizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	} else {
		for _, resource := range resources.Items {
			if err := kubernetesProvider.RemoveRoleBinding(ctx, mizuResourcesNamespace, resource.Name); err != nil {
				resourceDesc := fmt.Sprintf("RoleBinding %s in namespace %s", resource.Name, mizuResourcesNamespace)
				handleDeletionError(err, resourceDesc, &leftoverResources)
			}
		}
	}

	if err := kubernetesProvider.RemovePod(ctx, mizuResourcesNamespace, kubernetes.ApiServerPodName); err != nil {
		resourceDesc := fmt.Sprintf("Pod %s in namespace %s", kubernetes.ApiServerPodName, mizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	return leftoverResources
}

func handleDeletionError(err error, resourceDesc string, leftoverResources *[]string) {
	logger.Log.Debugf("Error removing %s: %v", resourceDesc, errormessage.FormatError(err))
	*leftoverResources = append(*leftoverResources, resourceDesc)
}
