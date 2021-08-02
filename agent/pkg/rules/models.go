package rules

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/google/martian/har"
	"github.com/up9inc/mizu/shared"
	jsonpath "github.com/yalp/jsonpath"
)

type RulesMatched struct {
	Matched bool              `json:"matched"`
	Rule    shared.RulePolicy `json:"rule"`
}

func appendRulesMatched(rulesMatched []RulesMatched, matched bool, rule shared.RulePolicy) []RulesMatched {
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
	enforcePolicy, _ := shared.DecodeEnforcePolicy("/app/enforce-policy/enforce-policy.yaml")
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
