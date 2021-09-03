package har

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/martian/har"
	"github.com/up9inc/mizu/tap"
	"github.com/up9inc/mizu/tap/api"
)

type HarFile struct {
	file       *os.File
	entryCount int
}

func BuildCookies(rawCookies []interface{}) []har.Cookie {
	cookies := make([]har.Cookie, 0, len(rawCookies))

	for _, cookie := range rawCookies {
		c := cookie.(map[string]interface{})
		expiresStr := ""
		if c["expires"] != nil {
			expiresStr = c["expires"].(string)
		}
		expires, _ := time.Parse(time.RFC3339, expiresStr)
		httpOnly := false
		if c["httponly"] != nil {
			httpOnly, _ = strconv.ParseBool(c["httponly"].(string))
		}
		secure := false
		if c["secure"] != nil {
			secure, _ = strconv.ParseBool(c["secure"].(string))
		}
		path := ""
		if c["path"] != nil {
			path = c["path"].(string)
		}
		domain := ""
		if c["domain"] != nil {
			domain = c["domain"].(string)
		}

		cookies = append(cookies, har.Cookie{
			Name:        c["name"].(string),
			Value:       c["value"].(string),
			Path:        path,
			Domain:      domain,
			HTTPOnly:    httpOnly,
			Secure:      secure,
			Expires:     expires,
			Expires8601: expiresStr,
		})
	}

	return cookies
}

func BuildHeaders(rawHeaders []interface{}) ([]har.Header, string, string, string, string, string) {
	var host, scheme, authority, path, status string
	headers := make([]har.Header, 0, len(rawHeaders))

	for _, header := range rawHeaders {
		h := header.(map[string]interface{})

		headers = append(headers, har.Header{
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
			path = h["value"].(string)
		}
	}

	return headers, host, scheme, authority, path, status
}

func BuildPostParams(rawParams []interface{}) []har.Param {
	params := make([]har.Param, 0, len(rawParams))
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

		params = append(params, har.Param{
			Name:        name,
			Value:       value,
			Filename:    fileName,
			ContentType: contentType,
		})
	}

	return params
}

func NewRequest(request *api.GenericMessage) (harRequest *har.Request, err error) {
	reqDetails := request.Payload.(map[string]interface{})["details"].(map[string]interface{})

	headers, host, scheme, authority, path, _ := BuildHeaders(reqDetails["headers"].([]interface{}))
	cookies := make([]har.Cookie, 0) // BuildCookies(reqDetails["cookies"].([]interface{}))

	postData, _ := reqDetails["postData"].(map[string]interface{})
	mimeType, _ := postData["mimeType"]
	if mimeType == nil || len(mimeType.(string)) == 0 {
		mimeType = "text/html"
	}
	text, _ := postData["text"]
	postDataText := ""
	if text != nil {
		postDataText = text.(string)
	}

	queryString := make([]har.QueryString, 0)
	for _, _qs := range reqDetails["queryString"].([]interface{}) {
		qs := _qs.(map[string]interface{})
		queryString = append(queryString, har.QueryString{
			Name:  qs["name"].(string),
			Value: qs["value"].(string),
		})
	}

	url := fmt.Sprintf("http://%s%s", host, reqDetails["url"].(string))
	if strings.HasPrefix(mimeType.(string), "application/grpc") {
		url = fmt.Sprintf("%s://%s%s", scheme, authority, path)
	}

	var harParams []har.Param
	if postData["params"] != nil {
		harParams = BuildPostParams(postData["params"].([]interface{}))
	}

	harRequest = &har.Request{
		Method:      reqDetails["method"].(string),
		URL:         url,
		HTTPVersion: reqDetails["httpVersion"].(string),
		HeadersSize: -1,
		BodySize:    int64(bytes.NewBufferString(postDataText).Len()),
		QueryString: queryString,
		Headers:     headers,
		Cookies:     cookies,
		PostData: &har.PostData{
			MimeType: mimeType.(string),
			Params:   harParams,
			Text:     postDataText,
		},
	}

	return
}

func NewResponse(response *api.GenericMessage) (harResponse *har.Response, err error) {
	resDetails := response.Payload.(map[string]interface{})["details"].(map[string]interface{})

	headers, _, _, _, _, _status := BuildHeaders(resDetails["headers"].([]interface{}))
	cookies := make([]har.Cookie, 0) // BuildCookies(resDetails["cookies"].([]interface{}))

	content, _ := resDetails["content"].(map[string]interface{})
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
		Text:     []byte(bodyText),
		Size:     int64(len(bodyText)),
	}

	status := int(resDetails["status"].(float64))
	if strings.HasPrefix(mimeType.(string), "application/grpc") {
		status, err = strconv.Atoi(_status)
		if err != nil {
			tap.SilentError("convert-response-status-for-har", "Failed converting status to int %s (%v,%+v)", err, err, err)
			return nil, errors.New("failed converting response status to int for HAR")
		}
	}

	harResponse = &har.Response{
		HTTPVersion: resDetails["httpVersion"].(string),
		Status:      status,
		StatusText:  resDetails["statusText"].(string),
		HeadersSize: -1,
		BodySize:    int64(bytes.NewBufferString(bodyText).Len()),
		Headers:     headers,
		Cookies:     cookies,
		Content:     harContent,
	}
	return
}

func NewEntry(pair *api.RequestResponsePair) (*har.Entry, error) {
	harRequest, err := NewRequest(&pair.Request)
	if err != nil {
		tap.SilentError("convert-request-to-har", "Failed converting request to HAR %s (%v,%+v)", err, err, err)
		return nil, errors.New("failed converting request to HAR")
	}

	harResponse, err := NewResponse(&pair.Response)
	if err != nil {
		fmt.Printf("err: %+v\n", err)
		tap.SilentError("convert-response-to-har", "Failed converting response to HAR %s (%v,%+v)", err, err, err)
		return nil, errors.New("failed converting response to HAR")
	}

	totalTime := pair.Response.CaptureTime.Sub(pair.Request.CaptureTime).Round(time.Millisecond).Milliseconds()
	if totalTime < 1 {
		totalTime = 1
	}

	harEntry := har.Entry{
		StartedDateTime: pair.Request.CaptureTime,
		Time:            totalTime,
		Request:         harRequest,
		Response:        harResponse,
		Cache:           &har.Cache{},
		Timings: &har.Timings{
			Send:    -1,
			Wait:    -1,
			Receive: totalTime,
		},
	}

	return &harEntry, nil
}

func (f *HarFile) WriteEntry(harEntry *har.Entry) {
	harEntryJson, err := json.Marshal(harEntry)
	if err != nil {
		tap.SilentError("har-entry-marshal", "Failed converting har entry object to JSON%s (%v,%+v)", err, err, err)
		return
	}

	var separator string
	if f.GetEntryCount() > 0 {
		separator = ","
	} else {
		separator = ""
	}

	harEntryString := append([]byte(separator), harEntryJson...)

	if _, err := f.file.Write(harEntryString); err != nil {
		log.Panicf("Failed to write to output file: %s (%v,%+v)", err, err, err)
	}

	f.entryCount++
}

func (f *HarFile) GetEntryCount() int {
	return f.entryCount
}

func (f *HarFile) Close() {
	f.writeTrailer()

	err := f.file.Close()
	if err != nil {
		log.Panicf("Failed to close output file: %s (%v,%+v)", err, err, err)
	}
}

func (f *HarFile) writeHeader() {
	header := []byte(`{"log": {"version": "1.2", "creator": {"name": "Mizu", "version": "0.0.1"}, "entries": [`)
	if _, err := f.file.Write(header); err != nil {
		log.Panicf("Failed to write header to output file: %s (%v,%+v)", err, err, err)
	}
}

func (f *HarFile) writeTrailer() {
	trailer := []byte("]}}")
	if _, err := f.file.Write(trailer); err != nil {
		log.Panicf("Failed to write trailer to output file: %s (%v,%+v)", err, err, err)
	}
}
