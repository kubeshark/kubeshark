package mizu

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/up9inc/mizu/shared"
	core "k8s.io/api/core/v1"
)

type ControlSocket struct {
	connection *websocket.Conn
}

func CreateControlSocket(socketServerAddress string) (*ControlSocket, error) {
	connection, err := shared.ConnectToSocketServer(socketServerAddress)
	if err != nil {
		return nil, err
	} else {
		return &ControlSocket{connection: connection}, nil
	}
}

func (controlSocket *ControlSocket) SendNewTappedPodsListMessage(namespace string, pods []core.Pod) error {
	podInfos := make([]shared.PodInfo, 0)
	for _, pod := range pods {
		podInfos = append(podInfos, shared.PodInfo{Name: pod.Name})
	}
	tapStatus := shared.TapStatus{Namespace: namespace, Pods: podInfos}
	socketMessage := shared.MizuSocketMessage{MessageType: shared.TAPPING_STATUS_MESSAGE_TYPE, Data: tapStatus}

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
