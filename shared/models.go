package shared

import (
	"io/ioutil"
	"strings"

	"github.com/op/go-logging"
	"github.com/up9inc/mizu/logger"

	"gopkg.in/yaml.v3"
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

type MizuAgentConfig struct {
	MaxDBSizeBytes         int64         `json:"maxDBSizeBytes"`
	InsertionFilter        string        `json:"insertionFilter"`
	AgentImage             string        `json:"agentImage"`
	PullPolicy             string        `json:"pullPolicy"`
	LogLevel               logging.Level `json:"logLevel"`
	TapperResources        Resources     `json:"tapperResources"`
	MizuResourcesNamespace string        `json:"mizuResourceNamespace"`
	AgentDatabasePath      string        `json:"agentDatabasePath"`
	ServiceMap             bool          `json:"serviceMap"`
	OAS                    bool          `json:"oas"`
	Telemetry              bool          `json:"telemetry"`
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

type RulesPolicy struct {
	Rules []RulePolicy `yaml:"rules"`
}

type RulePolicy struct {
	Type         string `yaml:"type"`
	Service      string `yaml:"service"`
	Path         string `yaml:"path"`
	Method       string `yaml:"method"`
	Key          string `yaml:"key"`
	Value        string `yaml:"value"`
	ResponseTime int64  `yaml:"response-time"`
	Name         string `yaml:"name"`
}

type RulesMatched struct {
	Matched bool       `json:"matched"`
	Rule    RulePolicy `json:"rule"`
}

func (r *RulePolicy) validateType() bool {
	permitedTypes := []string{"json", "header", "slo"}
	_, found := Find(permitedTypes, r.Type)
	if !found {
		logger.Log.Errorf("Only json, header and slo types are supported on rule definition. This rule will be ignored. rule name: %s", r.Name)
		found = false
	}
	if strings.ToLower(r.Type) == "slo" {
		if r.ResponseTime <= 0 {
			logger.Log.Errorf("When rule type is slo, the field response-time should be specified and have a value >= 1. rule name: %s", r.Name)
			found = false
		}
	}
	return found
}

func (rules *RulesPolicy) ValidateRulesPolicy() []int {
	invalidIndex := make([]int, 0)
	for i := range rules.Rules {
		validated := rules.Rules[i].validateType()
		if !validated {
			invalidIndex = append(invalidIndex, i)
		}
	}
	return invalidIndex
}

func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func DecodeEnforcePolicy(path string) (RulesPolicy, error) {
	content, err := ioutil.ReadFile(path)
	enforcePolicy := RulesPolicy{}
	if err != nil {
		return enforcePolicy, err
	}
	err = yaml.Unmarshal([]byte(content), &enforcePolicy)
	if err != nil {
		return enforcePolicy, err
	}
	invalidIndex := enforcePolicy.ValidateRulesPolicy()
	var k = 0
	if len(invalidIndex) != 0 {
		for i, rule := range enforcePolicy.Rules {
			if !ContainsInt(invalidIndex, i) {
				enforcePolicy.Rules[k] = rule
				k++
			}
		}
		enforcePolicy.Rules = enforcePolicy.Rules[:k]
	}
	return enforcePolicy, nil
}
