package up9

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"github.com/google/martian/har"
	"github.com/romana/rlog"
	"github.com/up9inc/mizu/shared"
	tapApi "github.com/up9inc/mizu/tap/api"
	"io/ioutil"
	"log"
	"mizuserver/pkg/database"
	"mizuserver/pkg/utils"
	"net/http"
	"net/url"
	"strings"
	"time"
	harUtils "github.com/up9inc/mizu/shared/utils"
)

const (
	AnalyzeCheckSleepTime = 5 * time.Second
)

type GuestToken struct {
	Token string `json:"token"`
	Model string `json:"model"`
}

type ModelStatus struct {
	LastMajorGeneration float64 `json:"lastMajorGeneration"`
}

func getGuestToken(url string, target *GuestToken) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	rlog.Infof("Got token from the server, starting to json decode... status code: %v", resp.StatusCode)
	return json.NewDecoder(resp.Body).Decode(target)
}

func CreateAnonymousToken(envPrefix string) (*GuestToken, error) {
	tokenUrl := fmt.Sprintf("https://trcc.%s/anonymous/token", envPrefix)
	if strings.HasPrefix(envPrefix, "http") {
		tokenUrl = fmt.Sprintf("%s/api/token", envPrefix)
	}
	token := &GuestToken{}
	if err := getGuestToken(tokenUrl, token); err != nil {
		rlog.Infof("Failed to get token, %s", err)
		return nil, err
	}
	return token, nil
}

func GetRemoteUrl(analyzeDestination string, analyzeToken string) string {
	return fmt.Sprintf("https://%s/share/%s", analyzeDestination, analyzeToken)
}

func CheckIfModelReady(analyzeDestination string, analyzeModel string, analyzeToken string) bool {
	statusUrl, _ := url.Parse(fmt.Sprintf("https://trcc.%s/models/%s/status", analyzeDestination, analyzeModel))
	req := &http.Request{
		Method: http.MethodGet,
		URL:    statusUrl,
		Header: map[string][]string{
			"Content-Type": {"application/json"},
			"Guest-Auth":   {analyzeToken},
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

func GetTrafficDumpUrl(analyzeDestination string, analyzeModel string) *url.URL {
	strUrl := fmt.Sprintf("https://traffic.%s/dumpTrafficBulk/%s", analyzeDestination, analyzeModel)
	if strings.HasPrefix(analyzeDestination, "http") {
		strUrl = fmt.Sprintf("%s/api/workspace/dumpTrafficBulk", analyzeDestination)
	}
	postUrl, _ := url.Parse(strUrl)
	return postUrl
}

type AnalyzeInformation struct {
	IsAnalyzing        bool
	SentCount          int
	AnalyzedModel      string
	AnalyzeToken       string
	AnalyzeDestination string
}

func (info *AnalyzeInformation) Reset() {
	info.IsAnalyzing = false
	info.AnalyzedModel = ""
	info.AnalyzeToken = ""
	info.AnalyzeDestination = ""
	info.SentCount = 0
}

var analyzeInformation = &AnalyzeInformation{}

func GetAnalyzeInfo() *shared.AnalyzeStatus {
	return &shared.AnalyzeStatus{
		IsAnalyzing:   analyzeInformation.IsAnalyzing,
		RemoteUrl:     GetRemoteUrl(analyzeInformation.AnalyzeDestination, analyzeInformation.AnalyzeToken),
		IsRemoteReady: CheckIfModelReady(analyzeInformation.AnalyzeDestination, analyzeInformation.AnalyzedModel, analyzeInformation.AnalyzeToken),
		SentCount:     analyzeInformation.SentCount,
	}
}

func UploadEntriesImpl(token string, model string, envPrefix string, sleepIntervalSec int) {
	analyzeInformation.IsAnalyzing = true
	analyzeInformation.AnalyzedModel = model
	analyzeInformation.AnalyzeToken = token
	analyzeInformation.AnalyzeDestination = envPrefix
	analyzeInformation.SentCount = 0

	sleepTime := time.Second * time.Duration(sleepIntervalSec)

	var timestampFrom int64 = 0

	for {
		timestampTo := time.Now().UnixNano() / int64(time.Millisecond)
		rlog.Infof("Getting entries from %v, to %v\n", timestampFrom, timestampTo)
		protocolFilter := "http"
		entriesArray := database.GetEntriesFromDb(timestampFrom, timestampTo, &protocolFilter)

		if len(entriesArray) > 0 {
			result := make([]har.Entry, 0)
			for _, data := range entriesArray {
				var pair tapApi.RequestResponsePair
				if err := json.Unmarshal([]byte(data.Entry), &pair); err != nil {
					continue
				}
				harEntry, err := harUtils.NewEntry(&pair)
				if err != nil {
					continue
				}
				if data.ResolvedSource != "" {
					harEntry.Request.Headers = append(harEntry.Request.Headers, har.Header{Name: "x-mizu-source", Value: data.ResolvedSource})
				}
				if data.ResolvedDestination != "" {
					harEntry.Request.Headers = append(harEntry.Request.Headers, har.Header{Name: "x-mizu-destination", Value: data.ResolvedDestination})
					harEntry.Request.URL = utils.SetHostname(harEntry.Request.URL, data.ResolvedDestination)
				}
				result = append(result, *harEntry)
			}

			rlog.Infof("About to upload %v entries\n", len(result))

			body, jMarshalErr := json.Marshal(result)
			if jMarshalErr != nil {
				analyzeInformation.Reset()
				rlog.Infof("Stopping analyzing")
				log.Fatal(jMarshalErr)
			}

			var in bytes.Buffer
			w := zlib.NewWriter(&in)
			_, _ = w.Write(body)
			_ = w.Close()
			reqBody := ioutil.NopCloser(bytes.NewReader(in.Bytes()))

			req := &http.Request{
				Method: http.MethodPost,
				URL:    GetTrafficDumpUrl(envPrefix, model),
				Header: map[string][]string{
					"Content-Encoding": {"deflate"},
					"Content-Type":     {"application/octet-stream"},
					"Guest-Auth":       {token},
				},
				Body: reqBody,
			}

			if _, postErr := http.DefaultClient.Do(req); postErr != nil {
				analyzeInformation.Reset()
				rlog.Info("Stopping analyzing")
				log.Fatal(postErr)
			}
			analyzeInformation.SentCount += len(entriesArray)
			rlog.Infof("Finish uploading %v entries to %s\n", len(entriesArray), GetTrafficDumpUrl(envPrefix, model))

		} else {
			rlog.Infof("Nothing to upload")
		}

		rlog.Infof("Sleeping for %v...\n", sleepTime)
		time.Sleep(sleepTime)
		timestampFrom = timestampTo
	}
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
