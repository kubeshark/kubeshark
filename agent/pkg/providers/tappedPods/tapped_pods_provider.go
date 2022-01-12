package tappedPods

import (
	"github.com/up9inc/mizu/shared"
	"mizuserver/pkg/providers/tappersStatus"
	"strings"
)

var tappedPods []*shared.PodInfo

func Get() []*shared.PodInfo {
	return tappedPods
}

func Set(tappedPodsToSet []*shared.PodInfo) {
	tappedPods = tappedPodsToSet
}

func GetTappedPodsStatus() []shared.TappedPodStatus {
	tappedPodsStatus := make([]shared.TappedPodStatus, 0)
	for _, pod := range Get() {
		var status string
		if tapperStatus, ok := tappersStatus.Get()[pod.NodeName]; ok {
			status = strings.ToLower(tapperStatus.Status)
		}

		isTapped := status == "running"
		tappedPodsStatus = append(tappedPodsStatus, shared.TappedPodStatus{Name: pod.Name, Namespace: pod.Namespace, IsTapped: isTapped})
	}

	return tappedPodsStatus
}
