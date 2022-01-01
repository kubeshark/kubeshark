package utils

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	har "github.com/mrichman/hargo"
	"github.com/up9inc/mizu/shared/logger"
)

// Keep it because we might want cookies in the future
//func BuildCookies(rawCookies []interface{}) []har.Cookie {
//	cookies := make([]har.Cookie, 0, len(rawCookies))
//
//	for _, cookie := range rawCookies {
//		c := cookie.(map[string]interface{})
//		expiresStr := ""
//		if c["expires"] != nil {
//			expiresStr = c["expires"].(string)
//		}
//		expires, _ := time.Parse(time.RFC3339, expiresStr)
//		httpOnly := false
//		if c["httponly"] != nil {
//			httpOnly, _ = strconv.ParseBool(c["httponly"].(string))
//		}
//		secure := false
//		if c["secure"] != nil {
//			secure, _ = strconv.ParseBool(c["secure"].(string))
//		}
//		path := ""
//		if c["path"] != nil {
//			path = c["path"].(string)
//		}
//		domain := ""
//		if c["domain"] != nil {
//			domain = c["domain"].(string)
//		}
//
//		cookies = append(cookies, har.Cookie{
//			Name:        c["name"].(string),
//			Value:       c["value"].(string),
//			Path:        path,
//			Domain:      domain,
//			HTTPOnly:    httpOnly,
//			Secure:      secure,
//			Expires:     expires,
//			Expires8601: expiresStr,
//		})
//	}
//
//	return cookies
//}

func BuildHeaders(rawHeaders []interface{}) ([]har.NVP, string, string, string, string, string) {
	var host, scheme, authority, path, status string
	headers := make([]har.NVP, 0, len(rawHeaders))

	for _, header := range rawHeaders {
		h := header.(map[string]interface{})

		headers = append(headers, har.NVP{
			Name:  h["name"].(string),
			Value: h["value"].(string),
		})

		if h["name"] == "Host" {
			host = h["value"].(string)
		}
		if h["name"] == ":authority" {
			authority = h["value"].(string)
		}
		if h["name"] == ":scheme" {
			scheme = h["value"].(string)
		}
		if h["name"] == ":path" {
			path = h["value"].(string)
		}
		if h["name"] == ":status" {
			status = h["value"].(string)
		}
	}

	return headers, host, scheme, authority, path, status
}

func BuildPostParams(rawParams []interface{}) []har.PostParam {
	params := make([]har.PostParam, 0, len(rawParams))
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

		params = append(params, har.PostParam{
			Name:        name,
			Value:       value,
			FileName:    fileName,
			ContentType: contentType,
		})
	}

	return params
}

func NewRequest(request map[string]interface{}) (harRequest *har.Request, err error) {
	headers, host, scheme, authority, path, _ := BuildHeaders(request["_headers"].([]interface{}))
	cookies := make([]har.Cookie, 0) // BuildCookies(request["_cookies"].([]interface{}))

	postData, _ := request["postData"].(map[string]interface{})
	mimeType, _ := postData["mimeType"]
	if mimeType == nil || len(mimeType.(string)) == 0 {
		mimeType = "text/html"
	}
	text, _ := postData["text"]
	postDataText := ""
	if text != nil {
		postDataText = text.(string)
	}

	queryString := make([]har.NVP, 0)
	for _, _qs := range request["_queryString"].([]interface{}) {
		qs := _qs.(map[string]interface{})
		queryString = append(queryString, har.NVP{
			Name:  qs["name"].(string),
			Value: qs["value"].(string),
		})
	}

	url := fmt.Sprintf("http://%s%s", host, request["url"].(string))
	if strings.HasPrefix(mimeType.(string), "application/grpc") {
		url = fmt.Sprintf("%s://%s%s", scheme, authority, path)
	}

	harParams := make([]har.PostParam, 0)
	if postData["params"] != nil {
		harParams = BuildPostParams(postData["params"].([]interface{}))
	}

	harRequest = &har.Request{
		Method:      request["method"].(string),
		URL:         url,
		HTTPVersion: request["httpVersion"].(string),
		HeaderSize:  -1,
		BodySize:    bytes.NewBufferString(postDataText).Len(),
		QueryString: queryString,
		Headers:     headers,
		Cookies:     cookies,
		PostData: har.PostData{
			MimeType: mimeType.(string),
			Params:   harParams,
			Text:     postDataText,
		},
	}

	return
}

func NewResponse(response map[string]interface{}) (harResponse *har.Response, err error) {
	headers, _, _, _, _, _status := BuildHeaders(response["_headers"].([]interface{}))
	cookies := make([]har.Cookie, 0) // BuildCookies(response["_cookies"].([]interface{}))

	content, _ := response["content"].(map[string]interface{})
	mimeType, _ := content["mimeType"]
	if mimeType == nil || len(mimeType.(string)) == 0 {
		mimeType = "text/html"
	}
	encoding, _ := content["encoding"]
	text, _ := content["text"]
	bodyText := ""
	if text != nil {
		bodyText = text.(string)
	}

	harContent := &har.Content{
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

	harResponse = &har.Response{
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

func NewEntry(request map[string]interface{}, response map[string]interface{}, startTime time.Time, elapsedTime int) (*har.Entry, error) {
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

	harEntry := har.Entry{
		StartedDateTime: startTime.String(),
		Time:            float32(elapsedTime),
		Request:         *harRequest,
		Response:        *harResponse,
		Cache:           har.Cache{},
		PageTimings: har.PageTimings{
			Send:    -1,
			Wait:    -1,
			Receive: elapsedTime,
		},
	}

	return &harEntry, nil
}
