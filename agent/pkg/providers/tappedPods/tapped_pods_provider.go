package tappedPods

import (
	"os"
	"strings"
	"sync"

	"github.com/kubeshark/kubeshark/agent/pkg/providers/tappers"
	"github.com/kubeshark/kubeshark/agent/pkg/utils"
	"github.com/kubeshark/kubeshark/logger"
	"github.com/kubeshark/kubeshark/shared"
)

const FilePath = shared.DataDirPath + "tapped-pods.json"

var (
	lock                    = &sync.Mutex{}
	syncOnce                sync.Once
	tappedPods              []*shared.PodInfo
	nodeHostToTappedPodsMap shared.NodeToPodsMap
)

func Get() []*shared.PodInfo {
	syncOnce.Do(func() {
		if err := utils.ReadJsonFile(FilePath, &tappedPods); err != nil {
			if !os.IsNotExist(err) {
				logger.Log.Errorf("Error reading tapped pods from file, err: %v", err)
			}
		}
	})

	return tappedPods
}

func Set(tappedPodsToSet []*shared.PodInfo) {
	lock.Lock()
	defer lock.Unlock()

	tappedPods = tappedPodsToSet
	if err := utils.SaveJsonFile(FilePath, tappedPods); err != nil {
		logger.Log.Errorf("Error saving tapped pods, err: %v", err)
	}
}

func GetTappedPodsStatus() []shared.TappedPodStatus {
	tappedPodsStatus := make([]shared.TappedPodStatus, 0)
	tapperStatus := tappers.GetStatus()
	for _, pod := range Get() {
		var status string
		if tapperStatus, ok := tapperStatus[pod.NodeName]; ok {
			status = strings.ToLower(tapperStatus.Status)
		}

		isTapped := status == "running"
		tappedPodsStatus = append(tappedPodsStatus, shared.TappedPodStatus{Name: pod.Name, Namespace: pod.Namespace, IsTapped: isTapped})
	}

	return tappedPodsStatus
}

func SetNodeToTappedPodMap(nodeToTappedPodsMap shared.NodeToPodsMap) {
	summary := nodeToTappedPodsMap.Summary()
	logger.Log.Debugf("Setting node to tapped pods map to %v", summary)

	nodeHostToTappedPodsMap = nodeToTappedPodsMap
}

func GetNodeToTappedPodMap() shared.NodeToPodsMap {
	return nodeHostToTappedPodsMap
}
