package mizu

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/up9inc/mizu/shared"
	core "k8s.io/api/core/v1"
	"time"
)

type ControlSocket struct {
	connection *websocket.Conn
}

func CreateControlSocket(socketServerAddress string) (*ControlSocket, error) {
	connection, err := shared.ConnectToSocketServer(socketServerAddress, 30, 2 * time.Second, true)
	if err != nil {
		return nil, err
	} else {
		return &ControlSocket{connection: connection}, nil
	}
}

func (controlSocket *ControlSocket) SendNewTappedPodsListMessage(pods []core.Pod) error {
	podInfos := make([]shared.PodInfo, 0)
	for _, pod := range pods {
		podInfos = append(podInfos, shared.PodInfo{Name: pod.Name, Namespace: pod.Namespace})
	}
	tapStatus := shared.TapStatus{Pods: podInfos}
	socketMessage := shared.CreateWebSocketStatusMessage(tapStatus)

	jsonMessage, err := json.Marshal(socketMessage)
	if err != nil {
		return err
	}
	err = controlSocket.connection.WriteMessage(websocket.TextMessage, jsonMessage)
	if err != nil {
		return err
	}

	return nil
}
