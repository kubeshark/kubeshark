package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/google/martian/har"
)

type HTTPPayload struct {
	Type uint8
	Data interface{}
}

type HTTPPayloader interface {
	MarshalJSON() ([]byte, error)
}

type HTTPWrapper struct {
	Method  string      `json:"method"`
	Url     string      `json:"url"`
	Details interface{} `json:"details"`
}

func (h HTTPPayload) MarshalJSON() ([]byte, error) {
	switch h.Type {
	case TypeHttpRequest:
		harRequest, err := har.NewRequest(h.Data.(*http.Request), true)
		if err != nil {
			return nil, errors.New("Failed converting request to HAR")
		}
		sort.Slice(harRequest.Headers, func(i, j int) bool {
			if harRequest.Headers[i].Name < harRequest.Headers[j].Name {
				return true
			}
			if harRequest.Headers[i].Name > harRequest.Headers[j].Name {
				return false
			}
			return harRequest.Headers[i].Value < harRequest.Headers[j].Value
		})
		sort.Slice(harRequest.QueryString, func(i, j int) bool {
			if harRequest.QueryString[i].Name < harRequest.QueryString[j].Name {
				return true
			}
			if harRequest.QueryString[i].Name > harRequest.QueryString[j].Name {
				return false
			}
			return harRequest.QueryString[i].Value < harRequest.QueryString[j].Value
		})
		if harRequest.PostData != nil {
			sort.Slice(harRequest.PostData.Params, func(i, j int) bool {
				if harRequest.PostData.Params[i].Name < harRequest.PostData.Params[j].Name {
					return true
				}
				if harRequest.PostData.Params[i].Name > harRequest.PostData.Params[j].Name {
					return false
				}
				return harRequest.PostData.Params[i].Value < harRequest.PostData.Params[j].Value
			})
		}
		return json.Marshal(&HTTPWrapper{
			Method:  harRequest.Method,
			Url:     "",
			Details: harRequest,
		})
	case TypeHttpResponse:
		harResponse, err := har.NewResponse(h.Data.(*http.Response), true)
		if err != nil {
			return nil, errors.New("Failed converting response to HAR")
		}
		sort.Slice(harResponse.Headers, func(i, j int) bool {
			if harResponse.Headers[i].Name < harResponse.Headers[j].Name {
				return true
			}
			if harResponse.Headers[i].Name > harResponse.Headers[j].Name {
				return false
			}
			return harResponse.Headers[i].Value < harResponse.Headers[j].Value
		})
		sort.Slice(harResponse.Cookies, func(i, j int) bool {
			if harResponse.Cookies[i].Name < harResponse.Cookies[j].Name {
				return true
			}
			if harResponse.Cookies[i].Name > harResponse.Cookies[j].Name {
				return false
			}
			return harResponse.Cookies[i].Value < harResponse.Cookies[j].Value
		})
		return json.Marshal(&HTTPWrapper{
			Method:  "",
			Url:     "",
			Details: harResponse,
		})
	default:
		panic(fmt.Sprintf("HTTP payload cannot be marshaled: %v", h.Type))
	}
}
