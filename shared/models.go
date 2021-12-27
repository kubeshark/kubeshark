package shared

import (
	"io/ioutil"
	"strings"

	"github.com/op/go-logging"
	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/tap/api"
	v1 "k8s.io/api/core/v1"

	"gopkg.in/yaml.v3"
)

type WebSocketMessageType string

const (
	WebSocketMessageTypeEntry         WebSocketMessageType = "entry"
	WebSocketMessageTypeTappedEntry   WebSocketMessageType = "tappedEntry"
	WebSocketMessageTypeUpdateStatus  WebSocketMessageType = "status"
	WebSocketMessageTypeAnalyzeStatus WebSocketMessageType = "analyzeStatus"
	WebsocketMessageTypeOutboundLink  WebSocketMessageType = "outboundLink"
	WebSocketMessageTypeToast         WebSocketMessageType = "toast"
	WebSocketMessageTypeQueryMetadata WebSocketMessageType = "queryMetadata"
	WebSocketMessageTypeStartTime     WebSocketMessageType = "startTime"
	WebSocketMessageTypeTapConfig     WebSocketMessageType = "tapConfig"
)

type Resources struct {
	CpuLimit       string `yaml:"cpu-limit" default:"750m"`
	MemoryLimit    string `yaml:"memory-limit" default:"1Gi"`
	CpuRequests    string `yaml:"cpu-requests" default:"50m"`
	MemoryRequests string `yaml:"memory-requests" default:"50Mi"`
}

type MizuAgentConfig struct {
	TapTargetRegex          api.SerializableRegexp      `json:"tapTargetRegex"`
	MaxDBSizeBytes          int64                       `json:"maxDBSizeBytes"`
	DaemonMode              bool                        `json:"daemonMode"`
	TargetNamespaces        []string                    `json:"targetNamespaces"`
	AgentImage              string                      `json:"agentImage"`
	PullPolicy              string                      `json:"pullPolicy"`
	LogLevel                logging.Level               `json:"logLevel"`
	IgnoredUserAgents       []string                    `json:"ignoredUserAgents"`
	TapperResources         Resources                   `json:"tapperResources"`
	MizuResourcesNamespace  string                      `json:"mizuResourceNamespace"`
	MizuApiFilteringOptions api.TrafficFilteringOptions `json:"mizuApiFilteringOptions"`
	AgentDatabasePath       string                      `json:"agentDatabasePath"`
	Istio                   bool                        `json:"istio"`
}

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
	TappingStatus []TappedPodStatus `json:"tappingStatus"`
}

type WebSocketTapConfigMessage struct {
	*WebSocketMessageMetadata
	TapTargets []v1.Pod `json:"pods"`
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

type TapStatus struct {
	Pods []PodInfo `json:"pods"`
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

type SyncEntriesConfig struct {
	Token             string `json:"token"`
	Env               string `json:"env"`
	Workspace         string `json:"workspace"`
	UploadIntervalSec int    `json:"interval"`
}

func CreateWebSocketStatusMessage(tappedPodsStatus []TappedPodStatus) WebSocketStatusMessage {
	return WebSocketStatusMessage{
		WebSocketMessageMetadata: &WebSocketMessageMetadata{
			MessageType: WebSocketMessageTypeUpdateStatus,
		},
		TappingStatus: tappedPodsStatus,
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

type HealthResponse struct {
	TapStatus     TapStatus      `json:"tapStatus"`
	TappersCount  int            `json:"tappersCount"`
	TappersStatus []TapperStatus `json:"tappersStatus"`
}

type VersionResponse struct {
	SemVer string `json:"semver"`
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
