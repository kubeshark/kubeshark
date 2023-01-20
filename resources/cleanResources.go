package resources

import (
	"context"
	"fmt"

	"github.com/kubeshark/kubeshark/errormessage"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/misc"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/util/wait"
)

func CleanUpSelfResources(ctx context.Context, cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider, isNsRestrictedMode bool, selfResourcesNamespace string) {
	log.Warn().Msg(fmt.Sprintf("Removing %s resources...", misc.Software))

	var leftoverResources []string

	if isNsRestrictedMode {
		leftoverResources = cleanUpRestrictedMode(ctx, kubernetesProvider, selfResourcesNamespace)
	} else {
		leftoverResources = cleanUpNonRestrictedMode(ctx, cancel, kubernetesProvider, selfResourcesNamespace)
	}

	if len(leftoverResources) > 0 {
		errMsg := "Failed to remove the following resources."
		for _, resource := range leftoverResources {
			errMsg += "\n- " + resource
		}
		log.Error().Msg(fmt.Sprintf(utils.Red, errMsg))
	}
}

func cleanUpNonRestrictedMode(ctx context.Context, cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider, selfResourcesNamespace string) []string {
	leftoverResources := make([]string, 0)

	if err := kubernetesProvider.RemoveNamespace(ctx, selfResourcesNamespace); err != nil {
		resourceDesc := fmt.Sprintf("Namespace %s", selfResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	} else {
		defer waitUntilNamespaceDeleted(ctx, cancel, kubernetesProvider, selfResourcesNamespace)
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

func waitUntilNamespaceDeleted(ctx context.Context, cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider, selfResourcesNamespace string) {
	// Call cancel if a terminating signal was received. Allows user to skip the wait.
	go func() {
		utils.WaitForTermination(ctx, cancel)
	}()

	if err := kubernetesProvider.WaitUtilNamespaceDeleted(ctx, selfResourcesNamespace); err != nil {
		switch {
		case ctx.Err() == context.Canceled:
			log.Printf("Do nothing. User interrupted the wait")
			log.Warn().
				Str("namespace", selfResourcesNamespace).
				Msg("Did nothing. User interrupted the wait.")
		case err == wait.ErrWaitTimeout:
			log.Warn().
				Str("namespace", selfResourcesNamespace).
				Msg("Timed out while deleting the namespace.")
		default:
			log.Warn().
				Err(errormessage.FormatError(err)).
				Str("namespace", selfResourcesNamespace).
				Msg("Unknown error while deleting the namespace.")
		}
	}
}

func cleanUpRestrictedMode(ctx context.Context, kubernetesProvider *kubernetes.Provider, selfResourcesNamespace string) []string {
	leftoverResources := make([]string, 0)

	if err := kubernetesProvider.RemoveService(ctx, selfResourcesNamespace, kubernetes.FrontServiceName); err != nil {
		resourceDesc := fmt.Sprintf("Service %s in namespace %s", kubernetes.FrontServiceName, selfResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveService(ctx, selfResourcesNamespace, kubernetes.HubServiceName); err != nil {
		resourceDesc := fmt.Sprintf("Service %s in namespace %s", kubernetes.HubServiceName, selfResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveDaemonSet(ctx, selfResourcesNamespace, kubernetes.WorkerDaemonSetName); err != nil {
		resourceDesc := fmt.Sprintf("DaemonSet %s in namespace %s", kubernetes.WorkerDaemonSetName, selfResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if resources, err := kubernetesProvider.ListManagedServiceAccounts(ctx, selfResourcesNamespace); err != nil {
		resourceDesc := fmt.Sprintf("ServiceAccounts in namespace %s", selfResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	} else {
		for _, resource := range resources.Items {
			if err := kubernetesProvider.RemoveServiceAccount(ctx, selfResourcesNamespace, resource.Name); err != nil {
				resourceDesc := fmt.Sprintf("ServiceAccount %s in namespace %s", resource.Name, selfResourcesNamespace)
				handleDeletionError(err, resourceDesc, &leftoverResources)
			}
		}
	}

	if resources, err := kubernetesProvider.ListManagedRoles(ctx, selfResourcesNamespace); err != nil {
		resourceDesc := fmt.Sprintf("Roles in namespace %s", selfResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	} else {
		for _, resource := range resources.Items {
			if err := kubernetesProvider.RemoveRole(ctx, selfResourcesNamespace, resource.Name); err != nil {
				resourceDesc := fmt.Sprintf("Role %s in namespace %s", resource.Name, selfResourcesNamespace)
				handleDeletionError(err, resourceDesc, &leftoverResources)
			}
		}
	}

	if resources, err := kubernetesProvider.ListManagedRoleBindings(ctx, selfResourcesNamespace); err != nil {
		resourceDesc := fmt.Sprintf("RoleBindings in namespace %s", selfResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	} else {
		for _, resource := range resources.Items {
			if err := kubernetesProvider.RemoveRoleBinding(ctx, selfResourcesNamespace, resource.Name); err != nil {
				resourceDesc := fmt.Sprintf("RoleBinding %s in namespace %s", resource.Name, selfResourcesNamespace)
				handleDeletionError(err, resourceDesc, &leftoverResources)
			}
		}
	}

	if err := kubernetesProvider.RemovePod(ctx, selfResourcesNamespace, kubernetes.HubPodName); err != nil {
		resourceDesc := fmt.Sprintf("Pod %s in namespace %s", kubernetes.HubPodName, selfResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemovePod(ctx, selfResourcesNamespace, kubernetes.FrontPodName); err != nil {
		resourceDesc := fmt.Sprintf("Pod %s in namespace %s", kubernetes.FrontPodName, selfResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	return leftoverResources
}

func handleDeletionError(err error, resourceDesc string, leftoverResources *[]string) {
	log.Warn().Err(errormessage.FormatError(err)).Msg(fmt.Sprintf("Error while removing %s", resourceDesc))
	*leftoverResources = append(*leftoverResources, resourceDesc)
}
