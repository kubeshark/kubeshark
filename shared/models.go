package shared

import (
	"io/ioutil"
	"log"
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
		log.Printf("Error: %s. ", r.Name)
		log.Printf("Only json, header and slo types are supported on rule definition. This rule will be ignored\n")
		found = false
	}
	if strings.ToLower(r.Type) == "slo" {
		if r.ResponseTime <= 0 {
			log.Printf("Error: %s. ", r.Name)
			log.Printf("When type=slo, the field response-time should be specified and have a value >= 1\n\n")
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
			if !contains(invalidIndex, i) {
				enforcePolicy.Rules[k] = rule
				k++
			}
		}
		enforcePolicy.Rules = enforcePolicy.Rules[:k]
	}
	return enforcePolicy, nil
}

func contains(s []int, num int) bool {
	for _, v := range s {
		if v == num {
			return true
		}
	}

	return false
}
