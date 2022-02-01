package up9

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/up9inc/mizu/agent/pkg/har"
	"github.com/up9inc/mizu/agent/pkg/utils"

	basenine "github.com/up9inc/basenine/client/go"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
	tapApi "github.com/up9inc/mizu/tap/api"
)

const (
	AnalyzeCheckSleepTime = 5 * time.Second
	SentCountLogInterval  = 100
)

type GuestToken struct {
	Token string `json:"token"`
	Model string `json:"model"`
}

type ModelStatus struct {
	LastMajorGeneration float64 `json:"lastMajorGeneration"`
}

func GetRemoteUrl(analyzeDestination string, analyzeModel string, analyzeToken string, guestMode bool) string {
	if guestMode {
		return fmt.Sprintf("https://%s/share/%s", analyzeDestination, analyzeToken)
	}

	return fmt.Sprintf("https://%s/app/workspaces/%s", analyzeDestination, analyzeModel)
}

func CheckIfModelReady(analyzeDestination string, analyzeModel string, analyzeToken string, guestMode bool) bool {
	statusUrl, _ := url.Parse(fmt.Sprintf("https://trcc.%s/models/%s/status", analyzeDestination, analyzeModel))

	authHeader := getAuthHeader(guestMode)
	req := &http.Request{
		Method: http.MethodGet,
		URL:    statusUrl,
		Header: map[string][]string{
			"Content-Type": {"application/json"},
			authHeader:     {analyzeToken},
		},
	}
	statusResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer statusResp.Body.Close()

	target := &ModelStatus{}
	_ = json.NewDecoder(statusResp.Body).Decode(&target)

	return target.LastMajorGeneration > 0
}

func getAuthHeader(guestMode bool) string {
	if guestMode {
		return "Guest-Auth"
	}

	return "Authorization"
}

func GetTrafficDumpUrl(analyzeDestination string, analyzeModel string) *url.URL {
	strUrl := fmt.Sprintf("https://traffic.%s/dumpTrafficBulk/%s", analyzeDestination, analyzeModel)
	postUrl, _ := url.Parse(strUrl)
	return postUrl
}

type AnalyzeInformation struct {
	IsAnalyzing        bool
	GuestMode          bool
	SentCount          int
	AnalyzedModel      string
	AnalyzeToken       string
	AnalyzeDestination string
}

func (info *AnalyzeInformation) Reset() {
	info.IsAnalyzing = false
	info.GuestMode = true
	info.AnalyzedModel = ""
	info.AnalyzeToken = ""
	info.AnalyzeDestination = ""
	info.SentCount = 0
}

var analyzeInformation = &AnalyzeInformation{}

func GetAnalyzeInfo() *shared.AnalyzeStatus {
	return &shared.AnalyzeStatus{
		IsAnalyzing:   analyzeInformation.IsAnalyzing,
		RemoteUrl:     GetRemoteUrl(analyzeInformation.AnalyzeDestination, analyzeInformation.AnalyzedModel, analyzeInformation.AnalyzeToken, analyzeInformation.GuestMode),
		IsRemoteReady: CheckIfModelReady(analyzeInformation.AnalyzeDestination, analyzeInformation.AnalyzedModel, analyzeInformation.AnalyzeToken, analyzeInformation.GuestMode),
		SentCount:     analyzeInformation.SentCount,
	}
}

func SyncEntries(syncEntriesConfig *shared.SyncEntriesConfig) error {
	logger.Log.Infof("Sync entries - started")

	var (
		token, model string
		guestMode    bool
	)
	if syncEntriesConfig.Token == "" {
		logger.Log.Infof("Sync entries - creating anonymous token. env %s", syncEntriesConfig.Env)
		guestToken, err := createAnonymousToken(syncEntriesConfig.Env)
		if err != nil {
			return fmt.Errorf("failed creating anonymous token, err: %v", err)
		}

		token = guestToken.Token
		model = guestToken.Model
		guestMode = true
	} else {
		token = fmt.Sprintf("bearer %s", syncEntriesConfig.Token)
		model = syncEntriesConfig.Workspace
		guestMode = false

		logger.Log.Infof("Sync entries - upserting model. env %s, model %s", syncEntriesConfig.Env, model)
		if err := upsertModel(token, model, syncEntriesConfig.Env); err != nil {
			return fmt.Errorf("failed upserting model, err: %v", err)
		}
	}

	modelRegex, _ := regexp.Compile("[A-Za-z0-9][-A-Za-z0-9_.]*[A-Za-z0-9]+$")
	if len(model) > 63 || !modelRegex.MatchString(model) {
		return fmt.Errorf("invalid model name, model name: %s", model)
	}

	logger.Log.Infof("Sync entries - syncing. token: %s, model: %s, guest mode: %v", token, model, guestMode)
	go syncEntriesImpl(token, model, syncEntriesConfig.Env, syncEntriesConfig.UploadIntervalSec, guestMode)

	return nil
}

func upsertModel(token string, model string, envPrefix string) error {
	upsertModelUrl, _ := url.Parse(fmt.Sprintf("https://trcc.%s/models/%s", envPrefix, model))

	authHeader := getAuthHeader(false)
	req := &http.Request{
		Method: http.MethodPost,
		URL:    upsertModelUrl,
		Header: map[string][]string{
			authHeader: {token},
		},
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed request to upsert model, err: %v", err)
	}

	// In case the model is not created (not 201) and doesn't exists (not 409)
	if response.StatusCode != 201 && response.StatusCode != 409 {
		return fmt.Errorf("failed request to upsert model, status code: %v", response.StatusCode)
	}

	return nil
}

func createAnonymousToken(envPrefix string) (*GuestToken, error) {
	tokenUrl := fmt.Sprintf("https://trcc.%s/anonymous/token", envPrefix)
	if strings.HasPrefix(envPrefix, "http") {
		tokenUrl = fmt.Sprintf("%s/api/token", envPrefix)
	}
	token := &GuestToken{}
	if err := getGuestToken(tokenUrl, token); err != nil {
		logger.Log.Infof("Failed to get token, %s", err)
		return nil, err
	}
	return token, nil
}

func getGuestToken(url string, target *GuestToken) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	logger.Log.Infof("Got token from the server, starting to json decode... status code: %v", resp.StatusCode)
	return json.NewDecoder(resp.Body).Decode(target)
}

func syncEntriesImpl(token string, model string, envPrefix string, uploadIntervalSec int, guestMode bool) {
	analyzeInformation.IsAnalyzing = true
	analyzeInformation.GuestMode = guestMode
	analyzeInformation.AnalyzedModel = model
	analyzeInformation.AnalyzeToken = token
	analyzeInformation.AnalyzeDestination = envPrefix
	analyzeInformation.SentCount = 0

	// "http or grpc" filter indicates that we're only interested in HTTP and gRPC entries
	query := "http or grpc"

	logger.Log.Infof("Getting entries from the database")

	var connection *basenine.Connection
	var err error
	connection, err = basenine.NewConnection(shared.BasenineHost, shared.BaseninePort)
	if err != nil {
		panic(err)
	}

	data := make(chan []byte)
	meta := make(chan []byte)

	defer func() {
		data <- []byte(basenine.CloseChannel)
		meta <- []byte(basenine.CloseChannel)
		connection.Close()
	}()

	lastTimeSynced := time.Time{}

	batch := make([]har.Entry, 0)

	handleDataChannel := func(wg *sync.WaitGroup, connection *basenine.Connection, data chan []byte) {
		defer wg.Done()
		for {
			dataBytes := <-data

			if string(dataBytes) == basenine.CloseChannel {
				return
			}

			var dataMap map[string]interface{}
			err = json.Unmarshal(dataBytes, &dataMap)

			var entry tapApi.Entry
			if err := json.Unmarshal([]byte(dataBytes), &entry); err != nil {
				continue
			}
			harEntry, err := har.NewEntry(entry.Request, entry.Response, entry.StartTime, entry.ElapsedTime)
			if err != nil {
				continue
			}
			if entry.Source.Name != "" {
				harEntry.Request.Headers = append(harEntry.Request.Headers, har.Header{Name: "x-mizu-source", Value: entry.Source.Name})
			}
			if entry.Destination.Name != "" {
				harEntry.Request.Headers = append(harEntry.Request.Headers, har.Header{Name: "x-mizu-destination", Value: entry.Destination.Name})
				harEntry.Request.URL = utils.SetHostname(harEntry.Request.URL, entry.Destination.Name)
			}

			batch = append(batch, *harEntry)

			now := time.Now()
			if lastTimeSynced.Add(time.Duration(uploadIntervalSec) * time.Second).After(now) {
				continue
			}
			lastTimeSynced = now

			body, jMarshalErr := json.Marshal(batch)
			batchSize := len(batch)
			if jMarshalErr != nil {
				analyzeInformation.Reset()
				logger.Log.Infof("Stopping sync entries")
				logger.Log.Fatal(jMarshalErr)
			}
			batch = make([]har.Entry, 0)

			var in bytes.Buffer
			w := zlib.NewWriter(&in)
			_, _ = w.Write(body)
			_ = w.Close()
			reqBody := ioutil.NopCloser(bytes.NewReader(in.Bytes()))

			authHeader := getAuthHeader(guestMode)
			req := &http.Request{
				Method: http.MethodPost,
				URL:    GetTrafficDumpUrl(envPrefix, model),
				Header: map[string][]string{
					"Content-Encoding": {"deflate"},
					"Content-Type":     {"application/octet-stream"},
					authHeader:         {token},
				},
				Body: reqBody,
			}

			if _, postErr := http.DefaultClient.Do(req); postErr != nil {
				analyzeInformation.Reset()
				logger.Log.Info("Stopping sync entries")
				logger.Log.Fatal(postErr)
			}
			analyzeInformation.SentCount += batchSize

			if analyzeInformation.SentCount%SentCountLogInterval == 0 {
				logger.Log.Infof("Uploaded %v entries until now", analyzeInformation.SentCount)
			}
		}
	}

	handleMetaChannel := func(wg *sync.WaitGroup, connection *basenine.Connection, meta chan []byte) {
		defer wg.Done()
		for {
			metaBytes := <-meta

			if string(metaBytes) == basenine.CloseChannel {
				return
			}
		}
	}

	var wg sync.WaitGroup
	go handleDataChannel(&wg, connection, data)
	go handleMetaChannel(&wg, connection, meta)
	wg.Add(2)

	connection.Query(query, data, meta)

	wg.Wait()
}

func UpdateAnalyzeStatus(callback func(data []byte)) {
	for {
		if !analyzeInformation.IsAnalyzing {
			time.Sleep(AnalyzeCheckSleepTime)
			continue
		}
		analyzeStatus := GetAnalyzeInfo()
		socketMessage := shared.CreateWebSocketMessageTypeAnalyzeStatus(*analyzeStatus)

		jsonMessage, _ := json.Marshal(socketMessage)
		callback(jsonMessage)
		time.Sleep(AnalyzeCheckSleepTime)
	}
}
