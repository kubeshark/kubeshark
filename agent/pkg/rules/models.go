package rules

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"regexp"
	"strings"

	"github.com/google/martian/har"
	jsonpath "github.com/yalp/jsonpath"
	yaml "gopkg.in/yaml.v3"
)

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

func (r *RulePolicy) validateType() bool {
	permitedTypes := []string{"json", "header", "latency"}
	_, found := Find(permitedTypes, r.Type)
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

type RulesMatched struct {
	Matched bool       `json:"matched"`
	Rule    RulePolicy `json:"rule"`
}

func appendRulesMatched(rulesMatched []RulesMatched, matched bool, rule RulePolicy) []RulesMatched {
	return append(rulesMatched, RulesMatched{Matched: matched, Rule: rule})
}

func ValidatePath(URLFromRule string, URL string) bool {
	if URLFromRule != "" {
		matchPath, err := regexp.MatchString(URLFromRule, URL)
		if err != nil || !matchPath {
			return false
		}
	}
	return true
}

func ValidateService(serviceFromRule string, service string) bool {
	if serviceFromRule != "" {
		matchService, err := regexp.MatchString(serviceFromRule, service)
		if err != nil || !matchService {
			return false
		}
	}
	return true
}

func MatchRequestPolicy(harEntry har.Entry, service string) (int, []RulesMatched) {
	enforcePolicy, _ := DecodeEnforcePolicy()
	var resultPolicyToSend []RulesMatched
	for _, rule := range enforcePolicy.Rules {
		if !ValidatePath(rule.Path, harEntry.Request.URL) || !ValidateService(rule.Service, service) {
			continue
		}
		if rule.Type == "json" {
			var bodyJsonMap interface{}
			_ = json.Unmarshal(harEntry.Response.Content.Text, &bodyJsonMap)
			out, _ := jsonpath.Read(bodyJsonMap, rule.Key)
			var matchValue bool
			if out == nil {
				continue
			}
			if reflect.TypeOf(out).Kind() == reflect.String {
				matchValue, _ = regexp.MatchString(rule.Value, out.(string))
			} else {
				val := fmt.Sprint(out)
				matchValue, _ = regexp.MatchString(rule.Value, val)
			}
			resultPolicyToSend = appendRulesMatched(resultPolicyToSend, matchValue, rule)
		} else if rule.Type == "header" {
			for j := range harEntry.Response.Headers {
				matchKey, _ := regexp.MatchString(rule.Key, harEntry.Response.Headers[j].Name)
				if matchKey {
					matchValue, _ := regexp.MatchString(rule.Value, harEntry.Response.Headers[j].Value)
					resultPolicyToSend = appendRulesMatched(resultPolicyToSend, matchValue, rule)
				}
			}
		} else {
			resultPolicyToSend = appendRulesMatched(resultPolicyToSend, true, rule)
		}
	}
	return len(enforcePolicy.Rules), resultPolicyToSend
}

func PassedValidationRules(rulesMatched []RulesMatched, numberOfRules int) (bool, int64) {
	if len(rulesMatched) == 0 {
		return false, -1
	}
	for _, rule := range rulesMatched {
		if rule.Matched == false {
			return false, -1
		}
	}
	for _, rule := range rulesMatched {
		if strings.ToLower(rule.Rule.Type) == "latency" {
			return true, rule.Rule.Latency
		}
	}
	return true, -1
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
			enforcePolicy.RemoveRule(invalidIndex[i])
		}
	}
	return enforcePolicy, nil
}

type VersionResponse struct {
	SemVer string `json:"semver"`
}
