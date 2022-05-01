package rules

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/up9inc/mizu/agent/pkg/har"

	"github.com/up9inc/mizu/logger"

	"github.com/up9inc/mizu/shared"
	"github.com/yalp/jsonpath"
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

func MatchRequestPolicy(harEntry har.Entry, service string) (resultPolicyToSend []RulesMatched, isEnabled bool) {
	enforcePolicy, err := shared.DecodeEnforcePolicy(fmt.Sprintf("%s%s", shared.ConfigDirPath, shared.ValidationRulesFileName))
	if err == nil && len(enforcePolicy.Rules) > 0 {
		isEnabled = true
	}
	for _, rule := range enforcePolicy.Rules {
		if !ValidatePath(rule.Path, harEntry.Request.URL) || !ValidateService(rule.Service, service) {
			continue
		}
		if rule.Type == "json" {
			var bodyJsonMap interface{}
			contentTextDecoded, _ := base64.StdEncoding.DecodeString(string(harEntry.Response.Content.Text))
			if err := json.Unmarshal(contentTextDecoded, &bodyJsonMap); err != nil {
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
				logger.Log.Info(matchValue, rule.Value)
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
	return
}

func PassedValidationRules(rulesMatched []RulesMatched) (bool, int64, int) {
	var numberOfRulesMatched = len(rulesMatched)
	var responseTime int64 = -1

	if numberOfRulesMatched == 0 {
		return false, 0, numberOfRulesMatched
	}

	for _, rule := range rulesMatched {
		if !rule.Matched {
			return false, responseTime, numberOfRulesMatched
		} else {
			if strings.ToLower(rule.Rule.Type) == "slo" {
				if rule.Rule.ResponseTime < responseTime || responseTime == -1 {
					responseTime = rule.Rule.ResponseTime
				}
			}
		}
	}

	return true, responseTime, numberOfRulesMatched
}
