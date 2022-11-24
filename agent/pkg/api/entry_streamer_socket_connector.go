package api

import (
	"fmt"

	"github.com/kubeshark/kubeshark/agent/pkg/models"
	tapApi "github.com/kubeshark/worker/api"
	basenine "github.com/up9inc/basenine/client/go"
)

type EntryStreamerSocketConnector interface {
	SendEntry(socketId int, entry *tapApi.Entry, params *WebSocketParams) error
	SendMetadata(socketId int, metadata *basenine.Metadata) error
	SendToastError(socketId int, err error) error
	CleanupSocket(socketId int)
}

type DefaultEntryStreamerSocketConnector struct{}

func (e *DefaultEntryStreamerSocketConnector) SendEntry(socketId int, entry *tapApi.Entry, params *WebSocketParams) error {
	var message []byte
	if params.EnableFullEntries {
		message, _ = models.CreateFullEntryWebSocketMessage(entry)
	} else {
		protocol, ok := protocolsMap[entry.Protocol.ToString()]
		if !ok {
			return fmt.Errorf("protocol not found, protocol: %v", protocol)
		}

		extension, ok := extensionsMap[protocol.Name]
		if !ok {
			return fmt.Errorf("extension not found, extension: %v", protocol.Name)
		}

		base := extension.Dissector.Summarize(entry)
		message, _ = models.CreateBaseEntryWebSocketMessage(base)
	}

	if err := SendToSocket(socketId, message); err != nil {
		return err
	}

	return nil
}

func (e *DefaultEntryStreamerSocketConnector) SendMetadata(socketId int, metadata *basenine.Metadata) error {
	metadataBytes, _ := models.CreateWebsocketQueryMetadataMessage(metadata)
	if err := SendToSocket(socketId, metadataBytes); err != nil {
		return err
	}

	return nil
}

func (e *DefaultEntryStreamerSocketConnector) SendToastError(socketId int, err error) error {
	toastBytes, _ := models.CreateWebsocketToastMessage(&models.ToastMessage{
		Type:      "error",
		AutoClose: 5000,
		Text:      fmt.Sprintf("Syntax error: %s", err.Error()),
	})
	if err := SendToSocket(socketId, toastBytes); err != nil {
		return err
	}

	return nil
}

func (e *DefaultEntryStreamerSocketConnector) CleanupSocket(socketId int) {
	socketObj := connectedWebsockets[socketId]
	socketCleanup(socketId, socketObj)
}
