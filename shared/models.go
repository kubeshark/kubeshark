package shared

type WebSocketMessageType string

const (
	WebSocketMessageTypeEntry         WebSocketMessageType = "entry"
	WebSocketMessageTypeTappedEntry   WebSocketMessageType = "tappedEntry"
	WebSocketMessageTypeUpdateStatus  WebSocketMessageType = "status"
	WebSocketMessageTypeAnalyzeStatus WebSocketMessageType = "analyzeStatus"
)

type WebSocketMessageMetadata struct {
	MessageType WebSocketMessageType `json:"messageType,omitempty"`
}

type WebSocketAnalyzeStatusMessage struct {
	*WebSocketMessageMetadata
	AnalyzeStatus AnalyzeStatus `json:"analyzeStatus"`
}

type AnalyzeStatus struct {
	IsAnalyzing   bool   `json:"isAnalyzing"`
	RemoteUrl     string `json:"remoteUrl"`
	IsRemoteReady bool   `json:"isRemoteReady"`
	SentCount     int    `json:"sentCount"`
}

type WebSocketStatusMessage struct {
	*WebSocketMessageMetadata
	TappingStatus TapStatus `json:"tappingStatus"`
}

type TapStatus struct {
	Pods []PodInfo `json:"pods"`
}

type PodInfo struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

func CreateWebSocketStatusMessage(tappingStatus TapStatus) WebSocketStatusMessage {
	return WebSocketStatusMessage{
		WebSocketMessageMetadata: &WebSocketMessageMetadata{
			MessageType: WebSocketMessageTypeUpdateStatus,
		},
		TappingStatus: tappingStatus,
	}
}

func CreateWebSocketMessageTypeAnalyzeStatus(analyzeStatus AnalyzeStatus) WebSocketAnalyzeStatusMessage {
	return WebSocketAnalyzeStatusMessage{
		WebSocketMessageMetadata: &WebSocketMessageMetadata{
			MessageType: WebSocketMessageTypeAnalyzeStatus,
		},
		AnalyzeStatus: analyzeStatus,
	}
}

type TrafficFilteringOptions struct {
	PlainTextMaskingRegexes []*SerializableRegexp
	HideHealthChecks        bool
	DisableRedaction        bool
}

type VersionResponse struct {
	SemVer string `json:"semver"`
}

