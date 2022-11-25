package check

import (
	"context"
	"fmt"
	"log"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/uiUtils"
)

func KubernetesResources(ctx context.Context, kubernetesProvider *kubernetes.Provider) bool {
	log.Printf("\nk8s-components\n--------------------")

	exist, err := kubernetesProvider.DoesNamespaceExist(ctx, config.Config.KubesharkResourcesNamespace)
	allResourcesExist := checkResourceExist(config.Config.KubesharkResourcesNamespace, "namespace", exist, err)

	exist, err = kubernetesProvider.DoesConfigMapExist(ctx, config.Config.KubesharkResourcesNamespace, kubernetes.ConfigMapName)
	allResourcesExist = checkResourceExist(kubernetes.ConfigMapName, "config map", exist, err) && allResourcesExist

	exist, err = kubernetesProvider.DoesServiceAccountExist(ctx, config.Config.KubesharkResourcesNamespace, kubernetes.ServiceAccountName)
	allResourcesExist = checkResourceExist(kubernetes.ServiceAccountName, "service account", exist, err) && allResourcesExist

	if config.Config.IsNsRestrictedMode() {
		exist, err = kubernetesProvider.DoesRoleExist(ctx, config.Config.KubesharkResourcesNamespace, kubernetes.RoleName)
		allResourcesExist = checkResourceExist(kubernetes.RoleName, "role", exist, err) && allResourcesExist

		exist, err = kubernetesProvider.DoesRoleBindingExist(ctx, config.Config.KubesharkResourcesNamespace, kubernetes.RoleBindingName)
		allResourcesExist = checkResourceExist(kubernetes.RoleBindingName, "role binding", exist, err) && allResourcesExist
	} else {
		exist, err = kubernetesProvider.DoesClusterRoleExist(ctx, kubernetes.ClusterRoleName)
		allResourcesExist = checkResourceExist(kubernetes.ClusterRoleName, "cluster role", exist, err) && allResourcesExist

		exist, err = kubernetesProvider.DoesClusterRoleBindingExist(ctx, kubernetes.ClusterRoleBindingName)
		allResourcesExist = checkResourceExist(kubernetes.ClusterRoleBindingName, "cluster role binding", exist, err) && allResourcesExist
	}

	exist, err = kubernetesProvider.DoesServiceExist(ctx, config.Config.KubesharkResourcesNamespace, kubernetes.ApiServerPodName)
	allResourcesExist = checkResourceExist(kubernetes.ApiServerPodName, "service", exist, err) && allResourcesExist

	allResourcesExist = checkPodResourcesExist(ctx, kubernetesProvider) && allResourcesExist

	return allResourcesExist
}

func checkPodResourcesExist(ctx context.Context, kubernetesProvider *kubernetes.Provider) bool {
	if pods, err := kubernetesProvider.ListPodsByAppLabel(ctx, config.Config.KubesharkResourcesNamespace, kubernetes.ApiServerPodName); err != nil {
		log.Printf("%v error checking if '%v' pod is running, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), kubernetes.ApiServerPodName, err)
		return false
	} else if len(pods) == 0 {
		log.Printf("%v '%v' pod doesn't exist", fmt.Sprintf(uiUtils.Red, "✗"), kubernetes.ApiServerPodName)
		return false
	} else if !kubernetes.IsPodRunning(&pods[0]) {
		log.Printf("%v '%v' pod not running", fmt.Sprintf(uiUtils.Red, "✗"), kubernetes.ApiServerPodName)
		return false
	}

	log.Printf("%v '%v' pod running", fmt.Sprintf(uiUtils.Green, "√"), kubernetes.ApiServerPodName)

	if pods, err := kubernetesProvider.ListPodsByAppLabel(ctx, config.Config.KubesharkResourcesNamespace, kubernetes.TapperPodName); err != nil {
		log.Printf("%v error checking if '%v' pods are running, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), kubernetes.TapperPodName, err)
		return false
	} else {
		tappers := 0
		notRunningTappers := 0

		for _, pod := range pods {
			tappers += 1
			if !kubernetes.IsPodRunning(&pod) {
				notRunningTappers += 1
			}
		}

		if notRunningTappers > 0 {
			log.Printf("%v '%v' %v/%v pods are not running", fmt.Sprintf(uiUtils.Red, "✗"), kubernetes.TapperPodName, notRunningTappers, tappers)
			return false
		}

		log.Printf("%v '%v' %v pods running", fmt.Sprintf(uiUtils.Green, "√"), kubernetes.TapperPodName, tappers)
		return true
	}
}

func checkResourceExist(resourceName string, resourceType string, exist bool, err error) bool {
	if err != nil {
		log.Printf("%v error checking if '%v' %v exists, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), resourceName, resourceType, err)
		return false
	} else if !exist {
		log.Printf("%v '%v' %v doesn't exist", fmt.Sprintf(uiUtils.Red, "✗"), resourceName, resourceType)
		return false
	}

	log.Printf("%v '%v' %v exists", fmt.Sprintf(uiUtils.Green, "√"), resourceName, resourceType)
	return true
}
