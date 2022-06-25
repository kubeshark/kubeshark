package replay

import (
	"bytes"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/up9inc/mizu/agent/pkg/app"
	"github.com/up9inc/mizu/shared"
	tapApi "github.com/up9inc/mizu/tap/api"
)

var (
	inProcessRequestsLocker = sync.Mutex{}
	inProcessRequests       = 0
)

const (
	maxParallelAction             = 5
	timeoutForSingleActionSeconds = time.Second * 120
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

func ExecuteRequest(replayData *shared.ReplayDetails) (interface{}, error) {
	if canMakeRequest() {
		defer decrementCounter()
		client := &http.Client{
			Timeout: timeoutForSingleActionSeconds,
		}

		request, err := http.NewRequest(replayData.Method, replayData.Url, bytes.NewBufferString(replayData.Body))
		if err != nil {
			return nil, err
		}

		for headerKey, headerValue := range replayData.Headers {
			request.Header.Add(headerKey, headerValue)
		}
		request.Header.Add("x-mizu-correlation-id", uuid.New().String())
		response, requestErr := client.Do(request)

		if requestErr != nil {
			return nil, requestErr
		}

		//var entry *tapApi.Entry
		//bytes, err := basenine.Single(shared.BasenineHost, shared.BaseninePort, entryId, singleEntryRequest.Query)
		//if err != nil {
		//	return nil, err
		//}
		//err = json.Unmarshal(bytes, &entry)
		//if err != nil {
		//	return nil, errors.New(string(bytes))
		//}
		//
		//extension := app.ExtensionsMap[entry.Protocol.Name]
		//base := extension.Dissector.Summarize(entry)
		//var representation []byte
		//representation, err = extension.Dissector.Represent(entry.Request, entry.Response)
		//if err != nil {
		//	return nil, err
		//}

		return &tapApi.EntryWrapper{
			Protocol:       *app.ExtensionsMap["http"].Protocol,
			Representation: string(representation),
			Data:           entry,
			Base:           base,
			Rules:          nil,
			IsRulesEnabled: false,
		}, nil

	} else {
		return nil, errors.New("busy in too manu requests")
	}
}

func decrementCounter() {
	inProcessRequestsLocker.Lock()
	inProcessRequests--
	inProcessRequestsLocker.Unlock()
}
