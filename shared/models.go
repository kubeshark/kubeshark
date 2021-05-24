package shared

type ControlSocketMessageType string

const (
	TAPPING_STATUS_MESSAGE_TYPE ControlSocketMessageType = "tappingStatus"
)

type MizuSocketMessage struct {
	MessageType ControlSocketMessageType `json:"messageType"`
	Data        interface{} `json:"data"`
}

type TapStatus struct {
	Namespace string `json:"namespace"`
	Pods      []PodInfo `json:"pods"`
}

type PodInfo struct {
	Name string `json:"name"`
}
