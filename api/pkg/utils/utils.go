package utils

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/martian/har"
	"log"
	"mizuserver/pkg/models"
	"mizuserver/pkg/tap"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
)

// StartServer starts the server with a graceful shutdown
func StartServer(app *fiber.App) {
	signals := make(chan os.Signal, 2)
	signal.Notify(signals,
		os.Interrupt,  	  // this catch ctrl + c
		syscall.SIGTSTP,  // this catch ctrl + z
	)

	go func() {
		_ = <-signals
		fmt.Println("Shutting down...")
		_ = app.Shutdown()
	}()

	// Run server.
	if err := app.Listen(":8899"); err != nil {
		log.Printf("Oops... Server is not running! Reason: %v", err)
	}
}

func ReverseSlice(data interface{}) {
	value := reflect.ValueOf(data)
	valueLen := value.Len()
	for i := 0; i <= int((valueLen-1)/2); i++ {
		reverseIndex := valueLen - 1 - i
		tmp := value.Index(reverseIndex).Interface()
		value.Index(reverseIndex).Set(value.Index(i))
		value.Index(i).Set(reflect.ValueOf(tmp))
	}
}

func CheckErr(e error) {
	if e != nil {
		log.Printf("%v", e)
		//panic(e)
	}
}

func SetHostname(address, newHostname string) string {
	replacedUrl, err := url.Parse(address)
	if err != nil{
		log.Printf("error replacing hostname to %s in address %s, returning original %v",newHostname, address, err)
		return address
	}
	replacedUrl.Host = newHostname
	return replacedUrl.String()

}

func GetResolvedBaseEntry(entry models.MizuEntry) models.BaseEntryDetails {
	entryUrl := entry.Url
	service := entry.Service
	if entry.ResolvedDestination != nil {
		entryUrl = SetHostname(entryUrl, *entry.ResolvedDestination)
		service = SetHostname(service, *entry.ResolvedDestination)
	}
	return models.BaseEntryDetails{
		Id:         entry.EntryId,
		Url:        entryUrl,
		Service:    service,
		Path:       entry.Path,
		StatusCode: entry.Status,
		Method:     entry.Method,
		Timestamp:  entry.Timestamp,
		RequestSenderIp: entry.RequestSenderIp,
	}
}

func GetBytesFromStruct(v interface{}) []byte{
	a, _ := json.Marshal(v)
	return a
}

func FilterSensitiveInfoFromHarRequest(harOutputItem *tap.OutputChannelItem) {
	filterHarHeaders(harOutputItem.HarEntry.Request.Headers)
	filterHarHeaders(harOutputItem.HarEntry.Response.Headers)

	harOutputItem.HarEntry.Request.URL = filterUrl(harOutputItem.HarEntry.Request.URL)

	var requestJsonMap map[string] interface{}
	err := json.Unmarshal([]byte(harOutputItem.HarEntry.Request.PostData.Text) ,&requestJsonMap)
	if err == nil {
		filterJsonMap(requestJsonMap)
	}
	//
	//filterJsonMap(harOutputItem.HarEntry.Response.Content.Text)


	// filter url query params
	// filter bodies
}

func filterHarHeaders(headers []har.Header) {
	for _, header := range headers {
		if isFieldNameSensitive(header.Name) {
			header.Value = maskedFieldPlaceholderValue
		}
	}
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

func filterUrl(originalUrl string) string {
	parsedUrl, err := url.Parse(originalUrl)
	if err != nil {
		return originalUrl
	} else {
		if len(parsedUrl.RawQuery) > 0 {
			newQueryArgs := make([]string, 0)
			for urlQueryParamName, urlQueryParamValues := range parsedUrl.Query() {
				newValues := urlQueryParamValues
				if isFieldNameSensitive(urlQueryParamName) {
					newValues = []string {maskedFieldPlaceholderValue}
				}
				for value := range newValues {
					newQueryArgs = append(newQueryArgs, fmt.Sprintf("%s=%s", urlQueryParamName, value))
				}
			}

			parsedUrl.RawQuery = strings.Join(newQueryArgs, "&")
		}

		return parsedUrl.String()
	}
}
