package tappedPods

import "github.com/up9inc/mizu/shared"

var tappedPods []*shared.PodInfo

func Get() []*shared.PodInfo {
	return tappedPods
}

func Set(tappedPodsToSet []*shared.PodInfo) {
	tappedPods = tappedPodsToSet
}
