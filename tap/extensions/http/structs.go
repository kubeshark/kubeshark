package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/martian/har"
	"github.com/romana/rlog"
)

type HTTPPayload struct {
	Type string
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
	case "http_request":
		harRequest, err := har.NewRequest(h.Data.(*http.Request), false)
		if err != nil {
			rlog.Debugf("convert-request-to-har", "Failed converting request to HAR %s (%v,%+v)", err, err, err)
			return nil, errors.New("Failed converting request to HAR")
		}
		return json.Marshal(&HTTPWrapper{
			Method:  harRequest.Method,
			Url:     "",
			Details: harRequest,
		})
	case "http_response":
		harResponse, err := har.NewResponse(h.Data.(*http.Response), true)
		if err != nil {
			rlog.Debugf("convert-response-to-har", "Failed converting response to HAR %s (%v,%+v)", err, err, err)
			return nil, errors.New("Failed converting response to HAR")
		}
		return json.Marshal(&HTTPWrapper{
			Method:  "",
			Url:     "",
			Details: harResponse,
		})
	default:
		panic(fmt.Sprintf("HTTP payload cannot be marshaled: %s\n", h.Type))
	}
}
