package shared

import (
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v3"
)

type WebSocketMessageType string

const (
	WebSocketMessageTypeEntry         WebSocketMessageType = "entry"
	WebSocketMessageTypeTappedEntry   WebSocketMessageType = "tappedEntry"
	WebSocketMessageTypeUpdateStatus  WebSocketMessageType = "status"
	WebSocketMessageTypeAnalyzeStatus WebSocketMessageType = "analyzeStatus"
	WebsocketMessageTypeOutboundLink  WebSocketMessageType = "outboundLink"
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
	Pods     []PodInfo     `json:"pods"`
	TLSLinks []TLSLinkInfo `json:"tlsLinks"`
}

type PodInfo struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type TLSLinkInfo struct {
	SourceIP                string `json:"sourceIP"`
	DestinationAddress      string `json:"destinationAddress"`
	ResolvedDestinationName string `json:"resolvedDestinationName"`
	ResolvedSourceName      string `json:"resolvedSourceName"`
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
	HealthChecksUserAgentHeaders []string
	PlainTextMaskingRegexes      []*SerializableRegexp
	DisableRedaction             bool
}

type VersionResponse struct {
	SemVer string `json:"semver"`
}

type RulesPolicy struct {
	Rules []RulePolicy `yaml:"rules"`
}

type RulePolicy struct {
	Type    string `yaml:"type"`
	Service string `yaml:"service"`
	Path    string `yaml:"path"`
	Method  string `yaml:"method"`
	Key     string `yaml:"key"`
	Value   string `yaml:"value"`
	Latency int64  `yaml:"latency"`
	Name    string `yaml:"name"`
}


type RulesMatched struct {
	Matched bool              `json:"matched"`
	Rule    RulePolicy `json:"rule"`
}

func (r *RulePolicy) validateType() bool {
	permitedTypes := []string{"json", "header", "latency"}
	_, found := Find(permitedTypes, r.Type)
	if !found {
		fmt.Printf("\nRule with name %s will be ignored. Err: only json, header and latency types are supported on rule definition.\n", r.Name)
	}
	if strings.ToLower(r.Type) == "latency" {
		if r.Latency == 0 {
			fmt.Printf("\nRule with name %s will be ignored. Err: when type=latency, the field Latency should be specified and have a value >= 1\n\n", r.Name)
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

func (rules *RulesPolicy) RemoveRule(idx int) {
	rules.Rules = append(rules.Rules[:idx], rules.Rules[idx+1:]...)
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
	if len(invalidIndex) != 0 {
		for i := range invalidIndex {
			enforcePolicy.RemoveRule(invalidIndex[i])
		}
	}
	return enforcePolicy, nil
}
