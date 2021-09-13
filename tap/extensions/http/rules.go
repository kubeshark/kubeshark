package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/google/martian/har"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/tap/api"
	jsonpath "github.com/yalp/jsonpath"
)



// TODO: until we fixed the Rules feature
func NewApplicableRules(status bool, latency int64, number int) api.ApplicableRules {
	ar := api.ApplicableRules{}
	ar.Status = status
	ar.Latency = latency
	ar.NumberOfRules = number
	return ar
}

func appendRulesMatched(rulesMatched []shared.RulesMatched, matched bool, rule shared.RulePolicy) []shared.RulesMatched {
	return append(rulesMatched, shared.RulesMatched{Matched: matched, Rule: rule})
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

func MatchRequestPolicy(harEntry har.Entry, service string) (int, []shared.RulesMatched) {
	enforcePolicy, _ := shared.DecodeEnforcePolicy(fmt.Sprintf("%s/%s", shared.RulePolicyPath, shared.RulePolicyFileName))
	var resultPolicyToSend []shared.RulesMatched
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

func PassedValidationRules(rulesMatched []shared.RulesMatched, numberOfRules int) (bool, int64, int) {
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