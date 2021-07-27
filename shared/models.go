package shared

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"regexp"
	"strings"

	"github.com/google/martian/har"
	jsonpath "github.com/yalp/jsonpath"
	yaml "gopkg.in/yaml.v2"
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

func MatchRequestPolicy(harEntry har.Entry, service string) (int, []RulesMatched) {
	enforcePolicy, _ := DecodeEnforcePolicy()
	var resultPolicyToSend []RulesMatched
	for _, value := range enforcePolicy.Rules {
		if value.Type == "json" {
			var bodyJsonMap interface{}
			_ = json.Unmarshal(harEntry.Response.Content.Text, &bodyJsonMap)
			if !value.ValidatePath(harEntry.Request.URL) || !value.ValidateService(service) {
				continue
			}
			out, err := jsonpath.Read(bodyJsonMap, value.Key)
			if err != nil {
				continue
			}
			var matchValue bool
			if reflect.TypeOf(out).Kind() == reflect.String {
				matchValue, err = regexp.MatchString(value.Value, out.(string))
			} else {
				val := fmt.Sprint(out)
				matchValue, err = regexp.MatchString(value.Value, val)
			}
			var result RulesMatched
			resultPolicyToSend = result.ReturnRulesMatchedObject(value, matchValue, resultPolicyToSend)
		} else if value.Type == "header" {
			for j := range harEntry.Response.Headers {
				if !value.ValidatePath(harEntry.Request.URL) || !value.ValidateService(service) {
					continue
				}
				matchKey, _ := regexp.MatchString(value.Key, harEntry.Response.Headers[j].Name)
				if matchKey {
					matchValue, _ := regexp.MatchString(value.Value, harEntry.Response.Headers[j].Value)
					var result RulesMatched
					resultPolicyToSend = result.ReturnRulesMatchedObject(value, matchValue, resultPolicyToSend)
				}
			}
		} else {

			if !value.ValidatePath(harEntry.Request.URL) || !value.ValidateService(service) {
				continue
			}
			var result RulesMatched
			resultPolicyToSend = result.ReturnRulesMatchedObject(value, true, resultPolicyToSend)
		}
	}
	return len(enforcePolicy.Rules), resultPolicyToSend
}

func PassedValidationRules(rulesMatched []RulesMatched, numberOfRules int) string {
	if len(rulesMatched) == 0 {
		return ""
	}
	for _, rule := range rulesMatched {
		if rule.Matched == false {
			return "red"
		}
	}
	for _, rule := range rulesMatched {
		if strings.ToLower(rule.Rule.Type) == "latency" {
			return fmt.Sprint(rule.Rule.Latency)
		}
	}
	return "green"
}

func DecodeEnforcePolicy() (RulesPolicy, error) {
	content, err := ioutil.ReadFile("/app/enforce-policy/enforce-policy.yaml")
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
			fmt.Println("only json, header and latency types are supported on rule")
			enforcePolicy.RemoveNotValidPolicy(invalidIndex[i])
		}
	}
	return enforcePolicy, nil
}

type VersionResponse struct {
	SemVer string `json:"semver"`
}
