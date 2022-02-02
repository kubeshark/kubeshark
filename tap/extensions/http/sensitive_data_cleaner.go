package http

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/beevik/etree"
	"github.com/up9inc/mizu/tap/api"
)

const maskedFieldPlaceholderValue = "[REDACTED]"
const userAgent = "user-agent"

//these values MUST be all lower case and contain no `-` or `_` characters
var personallyIdentifiableDataFields = []string{"token", "authorization", "authentication", "cookie", "userid", "password",
	"username", "user", "key", "passcode", "pass", "auth", "authtoken", "jwt",
	"bearer", "clientid", "clientsecret", "redirecturi", "phonenumber",
	"zip", "zipcode", "address", "country", "firstname", "lastname",
	"middlename", "fname", "lname", "birthdate"}

func IsIgnoredUserAgent(item *api.OutputChannelItem, options *api.TrafficFilteringOptions) bool {
	if item.Protocol.Name != "http" {
		return false
	}

	request := item.Pair.Request.Payload.(api.HTTPPayload).Data.(*http.Request)

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

func FilterSensitiveData(item *api.OutputChannelItem, options *api.TrafficFilteringOptions) {
	request := item.Pair.Request.Payload.(api.HTTPPayload).Data.(*http.Request)
	response := item.Pair.Response.Payload.(api.HTTPPayload).Data.(*http.Response)

	filterHeaders(&request.Header)
	filterHeaders(&response.Header)
	filterUrl(request.URL)
	filterRequestBody(request, options)
	filterResponseBody(response, options)
}

func filterRequestBody(request *http.Request, options *api.TrafficFilteringOptions) {
	contenType := getContentTypeHeaderValue(request.Header)
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return
	}
	filteredBody, err := filterHttpBody([]byte(body), contenType, options)
	if err == nil {
		request.Body = ioutil.NopCloser(bytes.NewBuffer(filteredBody))
	} else {
		request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}
}

func filterResponseBody(response *http.Response, options *api.TrafficFilteringOptions) {
	contentType := getContentTypeHeaderValue(response.Header)
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	filteredBody, err := filterHttpBody([]byte(body), contentType, options)
	if err == nil {
		response.Body = ioutil.NopCloser(bytes.NewBuffer(filteredBody))
	} else {
		response.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}
}

func filterHeaders(headers *http.Header) {
	for key := range *headers {
		if strings.ToLower(key) == userAgent {
			continue
		}

		if strings.ToLower(key) == "cookie" {
			headers.Del(key)
		} else if isFieldNameSensitive(key) {
			headers.Set(key, maskedFieldPlaceholderValue)
		}
	}
}

func getContentTypeHeaderValue(headers http.Header) string {
	for key := range headers {
		if strings.ToLower(key) == "content-type" {
			return headers.Get(key)
		}
	}
	return ""
}

func isFieldNameSensitive(fieldName string) bool {
	if fieldName == ":authority" {
		return false
	}

	name := strings.ToLower(fieldName)
	name = strings.ReplaceAll(name, "_", "")
	name = strings.ReplaceAll(name, "-", "")
	name = strings.ReplaceAll(name, " ", "")

	for _, sensitiveField := range personallyIdentifiableDataFields {
		if strings.Contains(name, sensitiveField) {
			return true
		}
	}

	return false
}

func filterHttpBody(bytes []byte, contentType string, options *api.TrafficFilteringOptions) ([]byte, error) {
	mimeType := strings.Split(contentType, ";")[0]
	switch strings.ToLower(mimeType) {
	case "application/json":
		return filterJsonBody(bytes)
	case "text/html":
		fallthrough
	case "application/xhtml+xml":
		fallthrough
	case "text/xml":
		fallthrough
	case "application/xml":
		return filterXmlEtree(bytes)
	case "text/plain":
		if options != nil && options.PlainTextMaskingRegexes != nil {
			return filterPlainText(bytes, options), nil
		}
	}
	return bytes, nil
}

func filterPlainText(bytes []byte, options *api.TrafficFilteringOptions) []byte {
	for _, regex := range options.PlainTextMaskingRegexes {
		bytes = regex.ReplaceAll(bytes, []byte(maskedFieldPlaceholderValue))
	}
	return bytes
}

func filterXmlEtree(bytes []byte) ([]byte, error) {
	if !IsValidXML(bytes) {
		return nil, errors.New("Invalid XML")
	}
	xmlDoc := etree.NewDocument()
	err := xmlDoc.ReadFromBytes(bytes)
	if err != nil {
		return nil, err
	} else {
		filterXmlElement(xmlDoc.Root())
	}
	return xmlDoc.WriteToBytes()
}

func IsValidXML(data []byte) bool {
	return xml.Unmarshal(data, new(interface{})) == nil
}

func filterXmlElement(element *etree.Element) {
	for i, attribute := range element.Attr {
		if isFieldNameSensitive(attribute.Key) {
			element.Attr[i].Value = maskedFieldPlaceholderValue
		}
	}
	if element.ChildElements() == nil || len(element.ChildElements()) == 0 {
		if isFieldNameSensitive(element.Tag) {
			element.SetText(maskedFieldPlaceholderValue)
		}
	} else {
		for _, element := range element.ChildElements() {
			filterXmlElement(element)
		}
	}
}

func filterJsonBody(bytes []byte) ([]byte, error) {
	var bodyJsonMap map[string]interface{}
	err := json.Unmarshal(bytes, &bodyJsonMap)
	if err != nil {
		return nil, err
	}
	filterJsonMap(bodyJsonMap)
	return json.Marshal(bodyJsonMap)
}

func filterJsonMap(jsonMap map[string]interface{}) {
	for key, value := range jsonMap {
		// Do not replace nil values with maskedFieldPlaceholderValue
		if value == nil {
			continue
		}

		nestedMap, isNested := value.(map[string]interface{})
		if isNested {
			filterJsonMap(nestedMap)
		} else {
			if isFieldNameSensitive(key) {
				jsonMap[key] = maskedFieldPlaceholderValue
			}
		}
	}
}

func filterUrl(url *url.URL) {
	if len(url.RawQuery) > 0 {
		newQueryArgs := make([]string, 0)
		for urlQueryParamName, urlQueryParamValues := range url.Query() {
			newValues := urlQueryParamValues
			if isFieldNameSensitive(urlQueryParamName) {
				newValues = []string{maskedFieldPlaceholderValue}
			}
			for _, paramValue := range newValues {
				newQueryArgs = append(newQueryArgs, fmt.Sprintf("%s=%s", urlQueryParamName, paramValue))
			}
		}

		url.RawQuery = strings.Join(newQueryArgs, "&")
	}
}
