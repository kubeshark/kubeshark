package replay

import (
	"bytes"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/up9inc/mizu/shared"
	tapApi "github.com/up9inc/mizu/tap/api"
	mizuhttp "github.com/up9inc/mizu/tap/extensions/http"
)

func Test(t *testing.T) {
	client := &http.Client{
		Timeout: 1 * time.Second,
	}
	replayData := shared.ReplayDetails{
		Method: "GET",
		Url:    "http://httpbin.org/bla",
		Body:   "",
		Headers: map[string]string{
			"Content-type": "plain/text",
		},
	}

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

	var representation []byte
	representation, err = extension.Dissector.Represent(entry.Request, entry.Response)
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
}
