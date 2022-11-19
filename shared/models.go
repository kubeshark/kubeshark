package shared

import (
	"github.com/op/go-logging"

	v1 "k8s.io/api/core/v1"
)

type WebSocketMessageType string

const (
	WebSocketMessageTypeEntry            WebSocketMessageType = "entry"
	WebSocketMessageTypeFullEntry        WebSocketMessageType = "fullEntry"
	WebSocketMessageTypeTappedEntry      WebSocketMessageType = "tappedEntry"
	WebSocketMessageTypeUpdateStatus     WebSocketMessageType = "status"
	WebSocketMessageTypeUpdateTappedPods WebSocketMessageType = "tappedPods"
	WebSocketMessageTypeToast            WebSocketMessageType = "toast"
	WebSocketMessageTypeQueryMetadata    WebSocketMessageType = "queryMetadata"
	WebSocketMessageTypeStartTime        WebSocketMessageType = "startTime"
	WebSocketMessageTypeTapConfig        WebSocketMessageType = "tapConfig"
)

type Resources struct {
	CpuLimit       string `yaml:"cpu-limit" default:"750m"`
	MemoryLimit    string `yaml:"memory-limit" default:"1Gi"`
	CpuRequests    string `yaml:"cpu-requests" default:"50m"`
	MemoryRequests string `yaml:"memory-requests" default:"50Mi"`
}

type OASConfig struct {
	Enable        bool `yaml:"enabled" default:"true"`
	MaxExampleLen int  `yaml:"max-example-len" default:"10240"`
}

type KubesharkAgentConfig struct {
	MaxDBSizeBytes              int64         `json:"maxDBSizeBytes"`
	InsertionFilter             string        `json:"insertionFilter"`
	AgentImage                  string        `json:"agentImage"`
	PullPolicy                  string        `json:"pullPolicy"`
	LogLevel                    logging.Level `json:"logLevel"`
	TapperResources             Resources     `json:"tapperResources"`
	KubesharkResourcesNamespace string        `json:"kubesharkResourceNamespace"`
	AgentDatabasePath           string        `json:"agentDatabasePath"`
	ServiceMap                  bool          `json:"serviceMap"`
	OAS                         OASConfig     `json:"oas"`
}

type WebSocketMessageMetadata struct {
	MessageType WebSocketMessageType `json:"messageType,omitempty"`
}

type WebSocketStatusMessage struct {
	*WebSocketMessageMetadata
	TappingStatus []TappedPodStatus `json:"tappingStatus"`
}

type WebSocketTappedPodsMessage struct {
	*WebSocketMessageMetadata
	NodeToTappedPodMap NodeToPodsMap `json:"nodeToTappedPodMap"`
}

type WebSocketTapConfigMessage struct {
	*WebSocketMessageMetadata
	TapTargets []v1.Pod `json:"pods"`
}

type NodeToPodsMap map[string][]v1.Pod

func (np NodeToPodsMap) Summary() map[string][]string {
	summary := make(map[string][]string)
	for node, pods := range np {
		for _, pod := range pods {
			summary[node] = append(summary[node], pod.Namespace+"/"+pod.Name)
		}
	}

	return summary
}

type TapperStatus struct {
	TapperName string `json:"tapperName"`
	NodeName   string `json:"nodeName"`
	Status     string `json:"status"`
}

type TappedPodStatus struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	IsTapped  bool   `json:"isTapped"`
}

type PodInfo struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	NodeName  string `json:"nodeName"`
}

type TLSLinkInfo struct {
	SourceIP                string `json:"sourceIP"`
	DestinationAddress      string `json:"destinationAddress"`
	ResolvedDestinationName string `json:"resolvedDestinationName"`
	ResolvedSourceName      string `json:"resolvedSourceName"`
}

func CreateWebSocketStatusMessage(tappedPodsStatus []TappedPodStatus) WebSocketStatusMessage {
	return WebSocketStatusMessage{
		WebSocketMessageMetadata: &WebSocketMessageMetadata{
			MessageType: WebSocketMessageTypeUpdateStatus,
		},
		TappingStatus: tappedPodsStatus,
	}
}

func CreateWebSocketTappedPodsMessage(nodeToTappedPodMap NodeToPodsMap) WebSocketTappedPodsMessage {
	return WebSocketTappedPodsMessage{
		WebSocketMessageMetadata: &WebSocketMessageMetadata{
			MessageType: WebSocketMessageTypeUpdateTappedPods,
		},
		NodeToTappedPodMap: nodeToTappedPodMap,
	}
}

type HealthResponse struct {
	TappedPods            []*PodInfo      `json:"tappedPods"`
	ConnectedTappersCount int             `json:"connectedTappersCount"`
	TappersStatus         []*TapperStatus `json:"tappersStatus"`
}

type VersionResponse struct {
	Ver string `json:"ver"`
}
