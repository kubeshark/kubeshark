package shared

import (
	"regexp"
)

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
}

// Enforce Policy

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
	Latency int    `yaml:"latency"`
	Name    string `yaml:"name"`
}

func (r RulePolicy) validateType() bool {
	permitedTypes := []string{"json", "header", "latency"}
	_, found := Find(permitedTypes, r.Type)
	return found
}

func (rules RulesPolicy) ValidateRulesPolicy() []int {
	invalidIndex := make([]int, 0)
	for i := range rules.Rules {
		validated := rules.Rules[i].validateType()
		if !validated {
			invalidIndex = append(invalidIndex, i)
		}
	}
	return invalidIndex
}

func (rules *RulesPolicy) RemoveNotValidPolicy(idx int) {
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

type RulesMatched struct {
	Matched bool       `json:"matched"`
	Rule    RulePolicy `json:"rule"`
}

func (rm RulesMatched) ReturnRulesMatchedObject(value RulePolicy, matched bool, resultPolicyToSend []RulesMatched) (rulesMatched []RulesMatched) {
	if matched {
		rm.Matched = true
	} else {
		rm.Matched = false
	}
	rm.Rule = value
	resultPolicyToSend = append(resultPolicyToSend, rm)
	return resultPolicyToSend
}

func (rule RulePolicy) ValidatePath(URL string) bool {
	if rule.Path != "" {
		matchPath, err := regexp.MatchString(rule.Path, URL)
		if err != nil || !matchPath {
			return false
		}
	}
	return true
}

func (rule RulePolicy) ValidateService(service string) bool {
	if rule.Service != "" {
		matchService, err := regexp.MatchString(rule.Service, service)
		if err != nil || !matchService {
			return false
		}
	}
	return true
}
