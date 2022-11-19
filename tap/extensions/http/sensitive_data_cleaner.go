package http

import (
	"net/http"
	"strings"

	"github.com/kubeshark/kubeshark/tap/api"
)

const userAgent = "user-agent"

func IsIgnoredUserAgent(item *api.OutputChannelItem, options *api.TrafficFilteringOptions) bool {
	if item.Protocol.Name != "http" {
		return false
	}

	request := item.Pair.Request.Payload.(HTTPPayload).Data.(*http.Request)

	for headerKey, headerValues := range request.Header {
		if strings.ToLower(headerKey) == userAgent {
			for _, userAgent := range options.IgnoredUserAgents {
				for _, headerValue := range headerValues {
					if strings.Contains(strings.ToLower(headerValue), strings.ToLower(userAgent)) {
						return true
					}
				}
			}

			return false
		}
	}

	return false
}
