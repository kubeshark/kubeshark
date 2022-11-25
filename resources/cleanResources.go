package resources

import (
	"context"
	"fmt"
	"log"

	"github.com/kubeshark/kubeshark/errormessage"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/uiUtils"
	"github.com/kubeshark/kubeshark/utils"
	"k8s.io/apimachinery/pkg/util/wait"
)

func CleanUpKubesharkResources(ctx context.Context, cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider, isNsRestrictedMode bool, kubesharkResourcesNamespace string) {
	log.Printf("\nRemoving kubeshark resources")

	var leftoverResources []string

	if isNsRestrictedMode {
		leftoverResources = cleanUpRestrictedMode(ctx, kubernetesProvider, kubesharkResourcesNamespace)
	} else {
		leftoverResources = cleanUpNonRestrictedMode(ctx, cancel, kubernetesProvider, kubesharkResourcesNamespace)
	}

	if len(leftoverResources) > 0 {
		errMsg := "Failed to remove the following resources."
		for _, resource := range leftoverResources {
			errMsg += "\n- " + resource
		}
		log.Printf(uiUtils.Error, errMsg)
	}
}

func cleanUpNonRestrictedMode(ctx context.Context, cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider, kubesharkResourcesNamespace string) []string {
	leftoverResources := make([]string, 0)

	if err := kubernetesProvider.RemoveNamespace(ctx, kubesharkResourcesNamespace); err != nil {
		resourceDesc := fmt.Sprintf("Namespace %s", kubesharkResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	} else {
		defer waitUntilNamespaceDeleted(ctx, cancel, kubernetesProvider, kubesharkResourcesNamespace)
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

func waitUntilNamespaceDeleted(ctx context.Context, cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider, kubesharkResourcesNamespace string) {
	// Call cancel if a terminating signal was received. Allows user to skip the wait.
	go func() {
		utils.WaitForFinish(ctx, cancel)
	}()

	if err := kubernetesProvider.WaitUtilNamespaceDeleted(ctx, kubesharkResourcesNamespace); err != nil {
		switch {
		case ctx.Err() == context.Canceled:
			log.Printf("Do nothing. User interrupted the wait")
		case err == wait.ErrWaitTimeout:
			log.Printf(uiUtils.Error, fmt.Sprintf("Timeout while removing Namespace %s", kubesharkResourcesNamespace))
		default:
			log.Printf(uiUtils.Error, fmt.Sprintf("Error while waiting for Namespace %s to be deleted: %v", kubesharkResourcesNamespace, errormessage.FormatError(err)))
		}
	}
}

func cleanUpRestrictedMode(ctx context.Context, kubernetesProvider *kubernetes.Provider, kubesharkResourcesNamespace string) []string {
	leftoverResources := make([]string, 0)

	if err := kubernetesProvider.RemoveService(ctx, kubesharkResourcesNamespace, kubernetes.ApiServerPodName); err != nil {
		resourceDesc := fmt.Sprintf("Service %s in namespace %s", kubernetes.ApiServerPodName, kubesharkResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveDaemonSet(ctx, kubesharkResourcesNamespace, kubernetes.TapperDaemonSetName); err != nil {
		resourceDesc := fmt.Sprintf("DaemonSet %s in namespace %s", kubernetes.TapperDaemonSetName, kubesharkResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveConfigMap(ctx, kubesharkResourcesNamespace, kubernetes.ConfigMapName); err != nil {
		resourceDesc := fmt.Sprintf("ConfigMap %s in namespace %s", kubernetes.ConfigMapName, kubesharkResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if resources, err := kubernetesProvider.ListManagedServiceAccounts(ctx, kubesharkResourcesNamespace); err != nil {
		resourceDesc := fmt.Sprintf("ServiceAccounts in namespace %s", kubesharkResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	} else {
		for _, resource := range resources.Items {
			if err := kubernetesProvider.RemoveServiceAccount(ctx, kubesharkResourcesNamespace, resource.Name); err != nil {
				resourceDesc := fmt.Sprintf("ServiceAccount %s in namespace %s", resource.Name, kubesharkResourcesNamespace)
				handleDeletionError(err, resourceDesc, &leftoverResources)
			}
		}
	}

	if resources, err := kubernetesProvider.ListManagedRoles(ctx, kubesharkResourcesNamespace); err != nil {
		resourceDesc := fmt.Sprintf("Roles in namespace %s", kubesharkResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	} else {
		for _, resource := range resources.Items {
			if err := kubernetesProvider.RemoveRole(ctx, kubesharkResourcesNamespace, resource.Name); err != nil {
				resourceDesc := fmt.Sprintf("Role %s in namespace %s", resource.Name, kubesharkResourcesNamespace)
				handleDeletionError(err, resourceDesc, &leftoverResources)
			}
		}
	}

	if resources, err := kubernetesProvider.ListManagedRoleBindings(ctx, kubesharkResourcesNamespace); err != nil {
		resourceDesc := fmt.Sprintf("RoleBindings in namespace %s", kubesharkResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	} else {
		for _, resource := range resources.Items {
			if err := kubernetesProvider.RemoveRoleBinding(ctx, kubesharkResourcesNamespace, resource.Name); err != nil {
				resourceDesc := fmt.Sprintf("RoleBinding %s in namespace %s", resource.Name, kubesharkResourcesNamespace)
				handleDeletionError(err, resourceDesc, &leftoverResources)
			}
		}
	}

	if err := kubernetesProvider.RemovePod(ctx, kubesharkResourcesNamespace, kubernetes.ApiServerPodName); err != nil {
		resourceDesc := fmt.Sprintf("Pod %s in namespace %s", kubernetes.ApiServerPodName, kubesharkResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	return leftoverResources
}

func handleDeletionError(err error, resourceDesc string, leftoverResources *[]string) {
	log.Printf("Error removing %s: %v", resourceDesc, errormessage.FormatError(err))
	*leftoverResources = append(*leftoverResources, resourceDesc)
}
