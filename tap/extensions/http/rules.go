package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"regexp"
	"strings"

	"github.com/google/martian/har"
	"github.com/up9inc/mizu/tap/api"
	"gopkg.in/yaml.v3"

	jsonpath "github.com/yalp/jsonpath"
)

var RulePolicyPath                   = "/app/enforce-policy/"
var RulePolicyFileName               = "enforce-policy.yaml"

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


// TODO: until we fixed the Rules feature
func NewApplicableRules(status bool, latency int64, number int) api.ApplicableRules {
	ar := api.ApplicableRules{}
	ar.Status = status
	ar.Latency = latency
	ar.NumberOfRules = number
	return ar
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
	enforcePolicy, _ := DecodeEnforcePolicy(fmt.Sprintf("%s/%s", RulePolicyPath, RulePolicyFileName))
	var resultPolicyToSend []RulesMatched
	for _, rule := range enforcePolicy.Rules {
		if !ValidatePath(rule.Path, harEntry.Request.URL) || !ValidateService(rule.Service, service) {
			continue
		}
		if rule.Type == "json" {
			var bodyJsonMap interface{}
			if err := json.Unmarshal(harEntry.Response.Content.Text, &bodyJsonMap); err != nil {
				continue
			}
			out, err := jsonpath.Read(bodyJsonMap, rule.Key)
			if err != nil || out == nil {
				continue
			}
			var matchValue bool
			if reflect.TypeOf(out).Kind() == reflect.String {
				matchValue, err = regexp.MatchString(rule.Value, out.(string))
				if err != nil {
					continue
				}
			} else {
				val := fmt.Sprint(out)
				matchValue, err = regexp.MatchString(rule.Value, val)
				if err != nil {
					continue
				}
			}
			resultPolicyToSend = appendRulesMatched(resultPolicyToSend, matchValue, rule)
		} else if rule.Type == "header" {
			for j := range harEntry.Response.Headers {
				matchKey, err := regexp.MatchString(rule.Key, harEntry.Response.Headers[j].Name)
				if err != nil {
					continue
				}
				if matchKey {
					matchValue, err := regexp.MatchString(rule.Value, harEntry.Response.Headers[j].Value)
					if err != nil {
						continue
					}
					resultPolicyToSend = appendRulesMatched(resultPolicyToSend, matchValue, rule)
				}
			}
		} else {
			resultPolicyToSend = appendRulesMatched(resultPolicyToSend, true, rule)
		}
	}
	return len(enforcePolicy.Rules), resultPolicyToSend
}

func PassedValidationRules(rulesMatched []RulesMatched, numberOfRules int) (bool, int64, int) {
	if len(rulesMatched) == 0 {
		return false, 0, 0
	}
	for _, rule := range rulesMatched {
		if rule.Matched == false {
			return false, -1, len(rulesMatched)
		}
	}
	for _, rule := range rulesMatched {
		if strings.ToLower(rule.Rule.Type) == "latency" {
			return true, rule.Rule.Latency, len(rulesMatched)
		}
	}
	return true, -1, len(rulesMatched)
}

func RunValidationRulesState(harEntry har.Entry, service string) api.ApplicableRules {
	numberOfRules, resultPolicyToSend := MatchRequestPolicy(harEntry, service)
	statusPolicyToSend, latency, numberOfRules := PassedValidationRules(resultPolicyToSend, numberOfRules)
	ar := NewApplicableRules(statusPolicyToSend, latency, numberOfRules)
	return ar
}