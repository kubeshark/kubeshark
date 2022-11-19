package har

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kubeshark/kubeshark/logger"
)

func BuildHeaders(rawHeaders map[string]interface{}) ([]Header, string, string, string, string, string) {
	var host, scheme, authority, path, status string
	headers := make([]Header, 0, len(rawHeaders))

	for key, value := range rawHeaders {
		headers = append(headers, Header{
			Name:  key,
			Value: value.(string),
		})

		if key == "Host" {
			host = value.(string)
		}
		if key == ":authority" {
			authority = value.(string)
		}
		if key == ":scheme" {
			scheme = value.(string)
		}
		if key == ":path" {
			path = value.(string)
		}
		if key == ":status" {
			status = value.(string)
		}
	}

	return headers, host, scheme, authority, path, status
}

func BuildPostParams(rawParams []interface{}) []Param {
	params := make([]Param, 0, len(rawParams))
	for _, param := range rawParams {
		p := param.(map[string]interface{})
		name := ""
		if p["name"] != nil {
			name = p["name"].(string)
		}
		value := ""
		if p["value"] != nil {
			value = p["value"].(string)
		}
		fileName := ""
		if p["fileName"] != nil {
			fileName = p["fileName"].(string)
		}
		contentType := ""
		if p["contentType"] != nil {
			contentType = p["contentType"].(string)
		}

		params = append(params, Param{
			Name:        name,
			Value:       value,
			FileName:    fileName,
			ContentType: contentType,
		})
	}

	return params
}

func NewRequest(request map[string]interface{}) (harRequest *Request, err error) {
	headers, host, scheme, authority, path, _ := BuildHeaders(request["headers"].(map[string]interface{}))
	cookies := make([]Cookie, 0)

	postData, _ := request["postData"].(map[string]interface{})
	mimeType := postData["mimeType"]
	if mimeType == nil {
		mimeType = ""
	}
	text := postData["text"]
	postDataText := ""
	if text != nil {
		postDataText = text.(string)
	}

	queryString := make([]QueryString, 0)
	for key, value := range request["queryString"].(map[string]interface{}) {
		if valuesInterface, ok := value.([]interface{}); ok {
			for _, valueInterface := range valuesInterface {
				queryString = append(queryString, QueryString{
					Name:  key,
					Value: valueInterface.(string),
				})
			}
		} else {
			queryString = append(queryString, QueryString{
				Name:  key,
				Value: value.(string),
			})
		}
	}

	url := fmt.Sprintf("http://%s%s", host, request["url"].(string))
	if strings.HasPrefix(mimeType.(string), "application/grpc") {
		url = fmt.Sprintf("%s://%s%s", scheme, authority, path)
	}

	harParams := make([]Param, 0)
	if postData["params"] != nil {
		harParams = BuildPostParams(postData["params"].([]interface{}))
	}

	harRequest = &Request{
		Method:      request["method"].(string),
		URL:         url,
		HTTPVersion: request["httpVersion"].(string),
		HeaderSize:  -1,
		BodySize:    bytes.NewBufferString(postDataText).Len(),
		QueryString: queryString,
		Headers:     headers,
		Cookies:     cookies,
		PostData: PostData{
			MimeType: mimeType.(string),
			Params:   harParams,
			Text:     postDataText,
		},
	}

	return
}

func NewResponse(response map[string]interface{}) (harResponse *Response, err error) {
	headers, _, _, _, _, _status := BuildHeaders(response["headers"].(map[string]interface{}))
	cookies := make([]Cookie, 0)

	content, _ := response["content"].(map[string]interface{})
	mimeType := content["mimeType"]
	if mimeType == nil {
		mimeType = ""
	}
	encoding := content["encoding"]
	text := content["text"]
	bodyText := ""
	if text != nil {
		bodyText = text.(string)
	}

	harContent := &Content{
		Encoding: encoding.(string),
		MimeType: mimeType.(string),
		Text:     bodyText,
		Size:     len(bodyText),
	}

	status := int(response["status"].(float64))
	if strings.HasPrefix(mimeType.(string), "application/grpc") {
		if _status != "" {
			status, err = strconv.Atoi(_status)
		}
		if err != nil {
			logger.Log.Errorf("Failed converting status to int %s (%v,%+v)", err, err, err)
			return nil, errors.New("failed converting response status to int for HAR")
		}
	}

	harResponse = &Response{
		HTTPVersion: response["httpVersion"].(string),
		Status:      status,
		StatusText:  response["statusText"].(string),
		HeadersSize: -1,
		BodySize:    bytes.NewBufferString(bodyText).Len(),
		Headers:     headers,
		Cookies:     cookies,
		Content:     *harContent,
	}
	return
}

func NewEntry(request map[string]interface{}, response map[string]interface{}, startTime time.Time, elapsedTime int64) (*Entry, error) {
	harRequest, err := NewRequest(request)
	if err != nil {
		logger.Log.Errorf("Failed converting request to HAR %s (%v,%+v)", err, err, err)
		return nil, errors.New("failed converting request to HAR")
	}

	harResponse, err := NewResponse(response)
	if err != nil {
		logger.Log.Errorf("Failed converting response to HAR %s (%v,%+v)", err, err, err)
		return nil, errors.New("failed converting response to HAR")
	}

	if elapsedTime < 1 {
		elapsedTime = 1
	}

	harEntry := Entry{
		StartedDateTime: startTime.Format(time.RFC3339),
		Time:            int(elapsedTime),
		Request:         *harRequest,
		Response:        *harResponse,
		Cache:           Cache{},
		PageTimings: PageTimings{
			Send:    -1,
			Wait:    -1,
			Receive: int(elapsedTime),
		},
	}

	return &harEntry, nil
}
