package sensitiveDataFiltering

import (
	"encoding/json"
	"fmt"
	"github.com/up9inc/mizu/shared"
	"mizuserver/pkg/tap"
	"net/url"
	"strings"

	"github.com/beevik/etree"
	"github.com/google/martian/har"
)

func FilterSensitiveInfoFromHarRequest(harOutputItem *tap.OutputChannelItem, options *shared.TrafficFilteringOptions) {
	filterHarHeaders(harOutputItem.HarEntry.Request.Headers)
	filterHarHeaders(harOutputItem.HarEntry.Response.Headers)

	harOutputItem.HarEntry.Request.Cookies = nil
	harOutputItem.HarEntry.Response.Cookies = nil

	harOutputItem.HarEntry.Request.URL = filterUrl(harOutputItem.HarEntry.Request.URL)
	for i, queryString := range harOutputItem.HarEntry.Request.QueryString {
		if isFieldNameSensitive(queryString.Name) {
			harOutputItem.HarEntry.Request.QueryString[i].Value = maskedFieldPlaceholderValue
		}
	}

	if harOutputItem.HarEntry.Request.PostData != nil {
		requestContentType := getContentTypeHeaderValue(harOutputItem.HarEntry.Request.Headers)
		filteredRequestBody, err := filterHttpBody([]byte(harOutputItem.HarEntry.Request.PostData.Text), requestContentType, options)
		if err == nil {
			harOutputItem.HarEntry.Request.PostData.Text = string(filteredRequestBody)
		}
	}
	if harOutputItem.HarEntry.Response.Content != nil {
		responseContentType := getContentTypeHeaderValue(harOutputItem.HarEntry.Response.Headers)
		filteredResponseBody, err := filterHttpBody(harOutputItem.HarEntry.Response.Content.Text, responseContentType, options)
		if err == nil {
			harOutputItem.HarEntry.Response.Content.Text = filteredResponseBody
		}
	}
}

func filterHarHeaders(headers []har.Header) {
	for i, header := range headers {
		if isFieldNameSensitive(header.Name) {
			headers[i].Value = maskedFieldPlaceholderValue
		}
	}
}

func getContentTypeHeaderValue(headers []har.Header) string {
	for _, header := range headers {
		if strings.ToLower(header.Name) == "content-type" {
			return header.Value
		}
	}
	return ""
}

func isFieldNameSensitive(fieldName string) bool {
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

func filterHttpBody(bytes []byte, contentType string, options *shared.TrafficFilteringOptions) ([]byte, error) {
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

func filterPlainText(bytes []byte, options *shared.TrafficFilteringOptions) []byte {
	for _, regex := range options.PlainTextMaskingRegexes {
		bytes = regex.ReplaceAll(bytes, []byte(maskedFieldPlaceholderValue))
	}
	return bytes
}

func filterXmlEtree(bytes []byte) ([]byte, error) {
	xmlDoc := etree.NewDocument()
	err := xmlDoc.ReadFromBytes(bytes)
	if err != nil {
		return nil, err
	} else {
		filterXmlElement(xmlDoc.Root())
	}
	return xmlDoc.WriteToBytes()
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
	var bodyJsonMap map[string] interface{}
	err := json.Unmarshal(bytes ,&bodyJsonMap)
	if err != nil {
		return nil, err
	}
	filterJsonMap(bodyJsonMap)
	return json.Marshal(bodyJsonMap)
}

func filterJsonMap(jsonMap map[string] interface{}) {
	for key, value := range jsonMap {
		if value == nil {
			return
		}
		nestedMap, isNested := value.(map[string] interface{})
		if isNested {
			filterJsonMap(nestedMap)
		} else {
			if isFieldNameSensitive(key) {
				jsonMap[key] = maskedFieldPlaceholderValue
			}
		}
	}
}

// receives string representing url, returns string url without sensitive query param values (http://service/api?userId=bob&password=123&type=login -> http://service/api?userId=[REDACTED]&password=[REDACTED]&type=login)
func filterUrl(originalUrl string) string {
	parsedUrl, err := url.Parse(originalUrl)
	if err != nil {
		return fmt.Sprintf("http://%s", maskedFieldPlaceholderValue)
	} else {
		if len(parsedUrl.RawQuery) > 0 {
			newQueryArgs := make([]string, 0)
			for urlQueryParamName, urlQueryParamValues := range parsedUrl.Query() {
				newValues := urlQueryParamValues
				if isFieldNameSensitive(urlQueryParamName) {
					newValues = []string {maskedFieldPlaceholderValue}
				}
				for _, paramValue := range newValues {
					newQueryArgs = append(newQueryArgs, fmt.Sprintf("%s=%s", urlQueryParamName, paramValue))
				}
			}

			parsedUrl.RawQuery = strings.Join(newQueryArgs, "&")
		}

		return parsedUrl.String()
	}
}
