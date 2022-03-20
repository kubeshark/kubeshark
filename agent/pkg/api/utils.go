package api

import (
	"encoding/json"

	"github.com/up9inc/mizu/agent/pkg/providers/tappedPods"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/kubernetes"
	"github.com/up9inc/mizu/shared/logger"
)

func BroadcastTappedPodsStatus() {
	tappedPodsStatus := tappedPods.GetTappedPodsStatus()

	message := shared.CreateWebSocketStatusMessage(tappedPodsStatus)
	if jsonBytes, err := json.Marshal(message); err != nil {
		logger.Log.Errorf("Could not Marshal message %v", err)
	} else {
		BroadcastToBrowserClients(jsonBytes)
	}
}

func BroadcastTappedPodsToTappers(nodeToTappedPodMap kubernetes.NodeToPodsMap) {
	message := shared.CreateWebSocketTappedPodsMessage(nodeToTappedPodMap)
	if jsonBytes, err := json.Marshal(message); err != nil {
		logger.Log.Errorf("Could not Marshal message %v", err)
	} else {
		BroadcastToTapperClients(jsonBytes)
	}
}
