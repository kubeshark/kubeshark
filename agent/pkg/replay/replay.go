package replay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/up9inc/mizu/agent/pkg/app"
	tapApi "github.com/up9inc/mizu/tap/api"
	mizuhttp "github.com/up9inc/mizu/tap/extensions/http"
)

var (
	inProcessRequestsLocker = sync.Mutex{}
	inProcessRequests       = 0
)

const maxParallelAction = 5

type Details struct {
	Method  string            `json:"method"`
	Url     string            `json:"url"`
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers"`
}

type Response struct {
	Success      bool        `json:"status"`
	Data         interface{} `json:"data"`
	ErrorMessage string      `json:"errorMessage"`
}

func incrementCounter() bool {
	result := false
	inProcessRequestsLocker.Lock()
	if inProcessRequests < maxParallelAction {
		inProcessRequests++
		result = true
	}
	inProcessRequestsLocker.Unlock()
	return result
}

func decrementCounter() {
	inProcessRequestsLocker.Lock()
	inProcessRequests--
	inProcessRequestsLocker.Unlock()
}

func getEntryFromRequestResponse(extension *tapApi.Extension, request *http.Request, response *http.Response) *tapApi.Entry {
	captureTime := time.Now()

	itemTmp := tapApi.OutputChannelItem{
		Protocol: *extension.Protocol,
		ConnectionInfo: &tapApi.ConnectionInfo{
			ClientIP:   "",
			ClientPort: "1",
			ServerIP:   "",
			ServerPort: "1",
			IsOutgoing: false,
		},
		Capture:   "",
		Timestamp: time.Now().UnixMilli(),
		Pair: &tapApi.RequestResponsePair{
			Request: tapApi.GenericMessage{
				IsRequest:   true,
				CaptureTime: captureTime,
				CaptureSize: 0,
				Payload: &mizuhttp.HTTPPayload{
					Type: mizuhttp.TypeHttpRequest,
					Data: request,
				},
			},
			Response: tapApi.GenericMessage{
				IsRequest:   false,
				CaptureTime: captureTime,
				CaptureSize: 0,
				Payload: &mizuhttp.HTTPPayload{
					Type: mizuhttp.TypeHttpResponse,
					Data: response,
				},
			},
		},
	}

	// Analyze is expecting an item that's marshalled and unmarshalled
	itemMarshalled, err := json.Marshal(itemTmp)
	if err != nil {
		return nil
	}
	var finalItem *tapApi.OutputChannelItem
	if err := json.Unmarshal(itemMarshalled, &finalItem); err != nil {
		return nil
	}

	return extension.Dissector.Analyze(finalItem, "", "", "")
}

func ExecuteRequest(replayData *Details, timeout time.Duration) *Response {
	if incrementCounter() {
		defer decrementCounter()

		client := &http.Client{
			Timeout: timeout,
		}

		request, err := http.NewRequest(strings.ToUpper(replayData.Method), replayData.Url, bytes.NewBufferString(replayData.Body))
		if err != nil {
			return &Response{
				Success:      false,
				Data:         nil,
				ErrorMessage: err.Error(),
			}
		}

		for headerKey, headerValue := range replayData.Headers {
			request.Header.Add(headerKey, headerValue)
		}
		request.Header.Add("x-mizu", uuid.New().String())
		response, requestErr := client.Do(request)

		if requestErr != nil {
			return &Response{
				Success:      false,
				Data:         nil,
				ErrorMessage: requestErr.Error(),
			}
		}

		extension := app.ExtensionsMap["http"] // # TODO: maybe pass the extension to the function so it can be tested
		entry := getEntryFromRequestResponse(extension, request, response)
		base := extension.Dissector.Summarize(entry)
		var representation []byte

		// Represent is expecting an entry that's marshalled and unmarshalled
		entryMarshalled, err := json.Marshal(entry)
		if err != nil {
			return &Response{
				Success:      false,
				Data:         nil,
				ErrorMessage: err.Error(),
			}
		}
		var entryUnmarshalled *tapApi.Entry
		if err := json.Unmarshal(entryMarshalled, &entryUnmarshalled); err != nil {
			return &Response{
				Success:      false,
				Data:         nil,
				ErrorMessage: err.Error(),
			}
		}

		representation, err = extension.Dissector.Represent(entryUnmarshalled.Request, entryUnmarshalled.Response)
		if err != nil {
			return &Response{
				Success:      false,
				Data:         nil,
				ErrorMessage: err.Error(),
			}
		}

		return &Response{
			Success: true,
			Data: &tapApi.EntryWrapper{
				Protocol:       *extension.Protocol,
				Representation: string(representation),
				Data:           entryUnmarshalled,
				Base:           base,
			},
			ErrorMessage: "",
		}
	} else {
		return &Response{
			Success:      false,
			Data:         nil,
			ErrorMessage: fmt.Sprintf("reached threshold of %d requests", maxParallelAction),
		}
	}
}
