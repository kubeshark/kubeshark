package replay

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/up9inc/mizu/shared"
	tapApi "github.com/up9inc/mizu/tap/api"
	mizuhttp "github.com/up9inc/mizu/tap/extensions/http"
)

func Test(t *testing.T) {
	client := &http.Client{
		Timeout: timeoutForSingleAction,
	}
	replayData := shared.ReplayDetails{
		Method:  "GET",
		Url:     "http://google.com",
		Body:    "",
		Headers: map[string]string{},
	}
	request, err := http.NewRequest(replayData.Method, replayData.Url, bytes.NewBufferString(replayData.Body))
	if err != nil {
		t.Errorf("failed: %v, ", err)
	}

	for headerKey, headerValue := range replayData.Headers {
		request.Header.Add(headerKey, headerValue)
	}
	request.Header.Add("x-mizu-correlation-id", uuid.New().String())
	response, requestErr := client.Do(request)

	if requestErr != nil {
		t.Errorf("failed: %v, ", requestErr)
	}

	captureTime := time.Now()

	extensionHttp := &tapApi.Extension{}
	dissectorHttp := mizuhttp.NewDissector()
	dissectorHttp.Register(extensionHttp)
	extensionHttp.Dissector = dissectorHttp
	extension := extensionHttp

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
		ConnectionInfo: &tapApi.ConnectionInfo{
			ClientIP:   "",
			ClientPort: "1",
			ServerIP:   "",
			ServerPort: "1",
			IsOutgoing: false,
		},
		Protocol:  *extension.Protocol,
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

	entry := *extension.Dissector.Analyze(&item, "", "", "")
	base := extension.Dissector.Summarize(&entry)
	t.Logf("%+v", entry)
	t.Logf("%+v", base)

}
