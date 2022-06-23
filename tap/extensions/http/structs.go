package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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
		return json.Marshal(&HTTPWrapper{
			Method:  "",
			Url:     "",
			Details: harResponse,
		})
	default:
		panic(fmt.Sprintf("HTTP payload cannot be marshaled: %v", h.Type))
	}
}
