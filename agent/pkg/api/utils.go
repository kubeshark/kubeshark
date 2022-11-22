package api

import (
	"encoding/json"

	"github.com/kubeshark/kubeshark/agent/pkg/providers/tappedPods"
	"github.com/kubeshark/kubeshark/logger"
	"github.com/kubeshark/kubeshark/shared"
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

func SendTappedPods(socketId int, nodeToTappedPodMap shared.NodeToPodsMap) {
	message := shared.CreateWebSocketTappedPodsMessage(nodeToTappedPodMap)
	if jsonBytes, err := json.Marshal(message); err != nil {
		logger.Log.Errorf("Could not Marshal message %v", err)
	} else {
		if err := SendToSocket(socketId, jsonBytes); err != nil {
			logger.Log.Error(err)
		}
	}
}

func BroadcastTappedPodsToTappers(nodeToTappedPodMap shared.NodeToPodsMap) {
	message := shared.CreateWebSocketTappedPodsMessage(nodeToTappedPodMap)
	if jsonBytes, err := json.Marshal(message); err != nil {
		logger.Log.Errorf("Could not Marshal message %v", err)
	} else {
		BroadcastToTapperClients(jsonBytes)
	}
}
