package shared

type WebSocketMessageType string
const (
	WebSocketMessageTypeEntry        WebSocketMessageType = "entry"
	WebSocketMessageTypeTappedEntry        WebSocketMessageType = "tappedEntry"
	WebSocketMessageTypeUpdateStatus WebSocketMessageType = "status"
)

type WebSocketMessageMetadata struct {
	MessageType WebSocketMessageType `json:"messageType,omitempty"`
}

type WebSocketStatusMessage struct {
	*WebSocketMessageMetadata
	TappingStatus TapStatus `json:"tappingStatus"`
}

type TapStatus struct {
	Pods      []PodInfo `json:"pods"`
}

type PodInfo struct {
	Namespace string `json:"namespace"`
	Name string `json:"name"`
}

func CreateWebSocketStatusMessage(tappingStatus TapStatus) WebSocketStatusMessage {
	return WebSocketStatusMessage{
		WebSocketMessageMetadata: &WebSocketMessageMetadata{
			MessageType: WebSocketMessageTypeUpdateStatus,
		},
		TappingStatus: tappingStatus,
	}
}
