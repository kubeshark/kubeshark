package replay

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/up9inc/mizu/agent/pkg/app"
	"github.com/up9inc/mizu/shared"
	tapApi "github.com/up9inc/mizu/tap/api"
	mizuhttp "github.com/up9inc/mizu/tap/extensions/http"
)

var (
	inProcessRequestsLocker = sync.Mutex{}
	inProcessRequests       = 0
)

const (
	maxParallelAction      = 5
	timeoutForSingleAction = time.Second * 30
)

func canMakeRequest() bool {
	result := false
	inProcessRequestsLocker.Lock()
	if inProcessRequests < maxParallelAction {
		inProcessRequests++
		result = true
	}
	inProcessRequestsLocker.Unlock()
	return result
}

func getEntryFromRequestResponse(extension *tapApi.Extension, request *http.Request, response *http.Response) *tapApi.Entry {
	captureTime := time.Now()

	httpRequestWrapperBytes, _ := json.Marshal(&mizuhttp.HTTPPayload{
		Type: mizuhttp.TypeHttpRequest,
		Data: request,
	})
	var httpRequestWrapper map[string]interface{}
	_ = json.Unmarshal(httpRequestWrapperBytes, &httpRequestWrapper)

	httpResponseWrapperBytes, _ := json.Marshal(&mizuhttp.HTTPPayload{
		Type: mizuhttp.TypeHttpResponse,
		Data: response,
	})
	var httpResponseWrapper map[string]interface{}
	_ = json.Unmarshal(httpResponseWrapperBytes, &httpResponseWrapper)

	item := tapApi.OutputChannelItem{
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
				Payload:     httpRequestWrapper,
			},
			Response: tapApi.GenericMessage{
				IsRequest:   false,
				CaptureTime: captureTime,
				CaptureSize: 0,
				Payload:     httpResponseWrapper,
			},
		},
	}

	return extension.Dissector.Analyze(&item, "", "", "")
}

func ExecuteRequest(replayData *shared.ReplayDetails, resultChannel chan *shared.ReplayResponse) {
	if canMakeRequest() {
		defer decrementCounter()
		// Handle Panics
		defer func() {
			if err := recover(); err != nil {
				resultChannel <- &shared.ReplayResponse{
					Success:      false,
					Data:         nil,
					ErrorMessage: err.(error).Error(),
				}
			}
		}()

		client := &http.Client{
			Timeout: timeoutForSingleAction,
		}

		request, err := http.NewRequest(strings.ToUpper(replayData.Method), replayData.Url, bytes.NewBufferString(replayData.Body))
		if err != nil {
			resultChannel <- &shared.ReplayResponse{
				Success:      false,
				Data:         nil,
				ErrorMessage: err.Error(),
			}
			return
		}

		for headerKey, headerValue := range replayData.Headers {
			request.Header.Add(headerKey, headerValue)
		}
		request.Header.Add("x-mizu", uuid.New().String())
		response, requestErr := client.Do(request)

		if requestErr != nil {
			resultChannel <- &shared.ReplayResponse{
				Success:      false,
				Data:         nil,
				ErrorMessage: requestErr.Error(),
			}
			return
		}

		extension := app.ExtensionsMap["http"] // # TODO: maybe pass the extension to the function so it can be tested
		entry := getEntryFromRequestResponse(extension, request, response)
		base := extension.Dissector.Summarize(entry)
		var representation []byte
		representation, err = extension.Dissector.Represent(entry.Request, entry.Response)
		if err != nil {
			resultChannel <- &shared.ReplayResponse{
				Success:      false,
				Data:         nil,
				ErrorMessage: err.Error(),
			}
			return
		}

		resultChannel <- &shared.ReplayResponse{
			Success: true,
			Data: &tapApi.EntryWrapper{
				Protocol:       *extension.Protocol,
				Representation: string(representation),
				Data:           entry,
				Base:           base,
				Rules:          nil,
				IsRulesEnabled: false,
			},
			ErrorMessage: "",
		}
	} else {
		resultChannel <- &shared.ReplayResponse{
			Success:      false,
			Data:         nil,
			ErrorMessage: "busy in too many requests",
		}
	}
}

func decrementCounter() {
	inProcessRequestsLocker.Lock()
	inProcessRequests--
	inProcessRequestsLocker.Unlock()
}
