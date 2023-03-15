package kubernetes

import (
	"github.com/kubeshark/base/pkg/models"
	"github.com/kubeshark/kubeshark/config"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetNodeHostToTargetedPodsMap(targetedPods []core.Pod) models.NodeToPodsMap {
	nodeToTargetedPodsMap := make(models.NodeToPodsMap)
	for _, pod := range targetedPods {
		minimizedPod := getMinimizedPod(pod)

		existingList := nodeToTargetedPodsMap[pod.Spec.NodeName]
		if existingList == nil {
			nodeToTargetedPodsMap[pod.Spec.NodeName] = []core.Pod{minimizedPod}
		} else {
			nodeToTargetedPodsMap[pod.Spec.NodeName] = append(nodeToTargetedPodsMap[pod.Spec.NodeName], minimizedPod)
		}
	}
	return nodeToTargetedPodsMap
}

func getMinimizedPod(fullPod core.Pod) core.Pod {
	return core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fullPod.Name,
			Namespace: fullPod.Namespace,
		},
		Status: core.PodStatus{
			PodIP:             fullPod.Status.PodIP,
			ContainerStatuses: getMinimizedContainerStatuses(fullPod),
		},
	}
}

func getMinimizedContainerStatuses(fullPod core.Pod) []core.ContainerStatus {
	result := make([]core.ContainerStatus, len(fullPod.Status.ContainerStatuses))

	for i, container := range fullPod.Status.ContainerStatuses {
		result[i] = core.ContainerStatus{
			ContainerID: container.ContainerID,
		}
	}

	return result
}

func GetPodInfosForPods(pods []core.Pod) []*models.PodInfo {
	podInfos := make([]*models.PodInfo, 0)
	for _, pod := range pods {
		podInfos = append(podInfos, &models.PodInfo{Name: pod.Name, Namespace: pod.Namespace, NodeName: pod.Spec.NodeName})
	}
	return podInfos
}

func buildWithDefaultLabels(labels map[string]string, provider *Provider) map[string]string {
	labels["LabelManagedBy"] = provider.managedBy
	labels["LabelCreatedBy"] = provider.createdBy

	for k, v := range config.Config.ResourceLabels {
		labels[k] = v
	}

	return labels
}
