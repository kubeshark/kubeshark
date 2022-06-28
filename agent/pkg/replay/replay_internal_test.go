package replay

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"encoding/json"

	"github.com/google/uuid"
	tapApi "github.com/up9inc/mizu/tap/api"
	mizuhttp "github.com/up9inc/mizu/tap/extensions/http"
)

func TestValid(t *testing.T) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	tests := map[string]*Details{
		"40x": {
			Method:  "GET",
			Url:     "http://httpbin.org/status/404",
			Body:    "",
			Headers: map[string]string{},
		},
		"20x": {
			Method:  "GET",
			Url:     "http://httpbin.org/status/200",
			Body:    "",
			Headers: map[string]string{},
		},
		"50x": {
			Method:  "GET",
			Url:     "http://httpbin.org/status/500",
			Body:    "",
			Headers: map[string]string{},
		},
		// TODO: this should be fixes, currently not working because of header name with ":"
		//":path-header": {
		//	Method: "GET",
		//	Url:    "http://httpbin.org/get",
		//	Body:   "",
		//	Headers: map[string]string{
		//		":path": "/get",
		//	},
		// },
	}

	for testCaseName, replayData := range tests {
		t.Run(fmt.Sprintf("%+v", testCaseName), func(t *testing.T) {
			request, err := http.NewRequest(strings.ToUpper(replayData.Method), replayData.Url, bytes.NewBufferString(replayData.Body))
			if err != nil {
				t.Errorf("Error executing request")
			}

			for headerKey, headerValue := range replayData.Headers {
				request.Header.Add(headerKey, headerValue)
			}
			request.Header.Add("x-mizu", uuid.New().String())
			response, requestErr := client.Do(request)

			if requestErr != nil {
				t.Errorf("failed: %v, ", requestErr)
			}

			extensionHttp := &tapApi.Extension{}
			dissectorHttp := mizuhttp.NewDissector()
			dissectorHttp.Register(extensionHttp)
			extensionHttp.Dissector = dissectorHttp
			extension := extensionHttp

			entry := getEntryFromRequestResponse(extension, request, response)
			base := extension.Dissector.Summarize(entry)

			// Represent is expecting an entry that's marshalled and unmarshalled
			entryMarshalled, err := json.Marshal(entry)
			if err != nil {
				t.Errorf("failed marshaling entry: %v, ", err)
			}
			var entryUnmarshalled *tapApi.Entry
			if err := json.Unmarshal(entryMarshalled, &entryUnmarshalled); err != nil {
				t.Errorf("failed unmarshaling entry: %v, ", err)
			}

			var representation []byte
			representation, err = extension.Dissector.Represent(entryUnmarshalled.Request, entryUnmarshalled.Response)
			if err != nil {
				t.Errorf("failed: %v, ", err)
			}

			result := &tapApi.EntryWrapper{
				Protocol:       *extension.Protocol,
				Representation: string(representation),
				Data:           entry,
				Base:           base,
				Rules:          nil,
				IsRulesEnabled: false,
			}
			t.Logf("%+v", result)
			//data, _ := json.MarshalIndent(result, "", "  ")
			//t.Logf("%+v", string(data))
		})
	}
}
