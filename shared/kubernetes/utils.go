package kubernetes

import (
	"regexp"

	"github.com/up9inc/mizu/shared"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetNodeHostToTappedPodsMap(tappedPods []core.Pod) shared.NodeToPodsMap {
	nodeToTappedPodMap := make(shared.NodeToPodsMap)
	for _, pod := range tappedPods {
		minimizedPod := getMinimizedPod(pod)

		existingList := nodeToTappedPodMap[pod.Spec.NodeName]
		if existingList == nil {
			nodeToTappedPodMap[pod.Spec.NodeName] = []core.Pod{minimizedPod}
		} else {
			nodeToTappedPodMap[pod.Spec.NodeName] = append(nodeToTappedPodMap[pod.Spec.NodeName], minimizedPod)
		}
	}
	return nodeToTappedPodMap
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

func excludeMizuPods(pods []core.Pod) []core.Pod {
	mizuPrefixRegex := regexp.MustCompile("^" + MizuResourcesPrefix)

	nonMizuPods := make([]core.Pod, 0)
	for _, pod := range pods {
		if !mizuPrefixRegex.MatchString(pod.Name) {
			nonMizuPods = append(nonMizuPods, pod)
		}
	}

	return nonMizuPods
}

func getPodArrayDiff(oldPods []core.Pod, newPods []core.Pod) (added []core.Pod, removed []core.Pod) {
	added = getMissingPods(newPods, oldPods)
	removed = getMissingPods(oldPods, newPods)

	return added, removed
}

//returns pods present in pods1 array and missing in pods2 array
func getMissingPods(pods1 []core.Pod, pods2 []core.Pod) []core.Pod {
	missingPods := make([]core.Pod, 0)
	for _, pod1 := range pods1 {
		var found = false
		for _, pod2 := range pods2 {
			if pod1.UID == pod2.UID {
				found = true
				break
			}
		}
		if !found {
			missingPods = append(missingPods, pod1)
		}
	}
	return missingPods
}

func GetPodInfosForPods(pods []core.Pod) []*shared.PodInfo {
	podInfos := make([]*shared.PodInfo, 0)
	for _, pod := range pods {
		podInfos = append(podInfos, &shared.PodInfo{Name: pod.Name, Namespace: pod.Namespace, NodeName: pod.Spec.NodeName})
	}
	return podInfos
}
