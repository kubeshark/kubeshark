package status

import (
	"encoding/json"

	"github.com/up9inc/mizu/agent/pkg/api"
	"github.com/up9inc/mizu/agent/pkg/providers/tappedPods"

	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
)

func BroadcastTappedPodsStatus() {
	tappedPodsStatus := tappedPods.GetTappedPodsStatus()

	message := shared.CreateWebSocketStatusMessage(tappedPodsStatus)
	if jsonBytes, err := json.Marshal(message); err != nil {
		logger.Log.Errorf("Could not Marshal message %v", err)
	} else {
		api.BroadcastToBrowserClients(jsonBytes)
	}
}
