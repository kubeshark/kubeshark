package check

import (
	"context"
	"fmt"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/rs/zerolog/log"
)

func KubernetesResources(ctx context.Context, kubernetesProvider *kubernetes.Provider) bool {
	log.Info().Str("procedure", "k8s-components").Msg("Checking:")

	exist, err := kubernetesProvider.DoesNamespaceExist(ctx, config.Config.SelfNamespace)
	allResourcesExist := checkResourceExist(config.Config.SelfNamespace, "namespace", exist, err)

	exist, err = kubernetesProvider.DoesServiceAccountExist(ctx, config.Config.SelfNamespace, kubernetes.ServiceAccountName)
	allResourcesExist = checkResourceExist(kubernetes.ServiceAccountName, "service account", exist, err) && allResourcesExist

	if config.Config.IsNsRestrictedMode() {
		exist, err = kubernetesProvider.DoesRoleExist(ctx, config.Config.SelfNamespace, kubernetes.RoleName)
		allResourcesExist = checkResourceExist(kubernetes.RoleName, "role", exist, err) && allResourcesExist

		exist, err = kubernetesProvider.DoesRoleBindingExist(ctx, config.Config.SelfNamespace, kubernetes.RoleBindingName)
		allResourcesExist = checkResourceExist(kubernetes.RoleBindingName, "role binding", exist, err) && allResourcesExist
	} else {
		exist, err = kubernetesProvider.DoesClusterRoleExist(ctx, kubernetes.ClusterRoleName)
		allResourcesExist = checkResourceExist(kubernetes.ClusterRoleName, "cluster role", exist, err) && allResourcesExist

		exist, err = kubernetesProvider.DoesClusterRoleBindingExist(ctx, kubernetes.ClusterRoleBindingName)
		allResourcesExist = checkResourceExist(kubernetes.ClusterRoleBindingName, "cluster role binding", exist, err) && allResourcesExist
	}

	exist, err = kubernetesProvider.DoesServiceExist(ctx, config.Config.SelfNamespace, kubernetes.HubServiceName)
	allResourcesExist = checkResourceExist(kubernetes.HubServiceName, "service", exist, err) && allResourcesExist

	allResourcesExist = checkPodResourcesExist(ctx, kubernetesProvider) && allResourcesExist

	return allResourcesExist
}

func checkPodResourcesExist(ctx context.Context, kubernetesProvider *kubernetes.Provider) bool {
	if pods, err := kubernetesProvider.ListPodsByAppLabel(ctx, config.Config.SelfNamespace, kubernetes.HubPodName); err != nil {
		log.Error().
			Str("name", kubernetes.HubPodName).
			Err(err).
			Msg("While checking if pod is running!")
		return false
	} else if len(pods) == 0 {
		log.Error().
			Str("name", kubernetes.HubPodName).
			Msg("Pod doesn't exist!")
		return false
	} else if !kubernetes.IsPodRunning(&pods[0]) {
		log.Error().
			Str("name", kubernetes.HubPodName).
			Msg("Pod is not running!")
		return false
	}

	log.Info().
		Str("name", kubernetes.HubPodName).
		Msg("Pod is running.")

	if pods, err := kubernetesProvider.ListPodsByAppLabel(ctx, config.Config.SelfNamespace, kubernetes.WorkerPodName); err != nil {
		log.Error().
			Str("name", kubernetes.WorkerPodName).
			Err(err).
			Msg("While checking if pods are running!")
		return false
	} else {
		workers := 0
		notRunningWorkers := 0

		for _, pod := range pods {
			workers += 1
			if !kubernetes.IsPodRunning(&pod) {
				notRunningWorkers += 1
			}
		}

		if notRunningWorkers > 0 {
			log.Error().
				Str("name", kubernetes.WorkerPodName).
				Msg(fmt.Sprintf("%d/%d pods are not running!", notRunningWorkers, workers))
			return false
		}

		log.Info().
			Str("name", kubernetes.WorkerPodName).
			Msg(fmt.Sprintf("All %d pods are running.", workers))
		return true
	}
}

func checkResourceExist(resourceName string, resourceType string, exist bool, err error) bool {
	if err != nil {
		log.Error().
			Str("name", resourceName).
			Str("type", resourceType).
			Err(err).
			Msg("Checking if resource exists!")
		return false
	} else if !exist {
		log.Error().
			Str("name", resourceName).
			Str("type", resourceType).
			Msg("Resource doesn't exist!")
		return false
	}

	log.Info().
		Str("name", resourceName).
		Str("type", resourceType).
		Msg("Resource exist.")
	return true
}
