package up9

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"github.com/up9inc/mizu/shared"
	"io/ioutil"
	"log"
	"mizuserver/pkg/database"
	"net/http"
	"net/url"
	"time"
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

	return json.NewDecoder(resp.Body).Decode(target)
}

func CreateAnonymousToken(envPrefix string) (*GuestToken, error) {
	tokenUrl := fmt.Sprintf("https://trcc.%v/anonymous/token", envPrefix)
	token := &GuestToken{}
	if err := getGuestToken(tokenUrl, token); err != nil {
		fmt.Println(err)
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
	postUrl, _ := url.Parse(fmt.Sprintf("https://traffic.%s/dumpTrafficBulk/%s", analyzeDestination, analyzeModel))
	return postUrl
}

type AnalyzeInformation struct {
	IsAnalyzing        bool
	AnalyzedModel      string
	AnalyzeToken       string
	AnalyzeDestination string
}

func (info *AnalyzeInformation) Reset() {
	info.IsAnalyzing = false
	info.AnalyzedModel = ""
	info.AnalyzeToken = ""
	info.AnalyzeDestination = ""
}

var analyzeInformation = &AnalyzeInformation{}

func GetAnalyzeInfo() *shared.AnalyzeStatus {
	return &shared.AnalyzeStatus{
		IsAnalyzing:   analyzeInformation.IsAnalyzing,
		RemoteUrl:     GetRemoteUrl(analyzeInformation.AnalyzeDestination, analyzeInformation.AnalyzeToken),
		IsRemoteReady: CheckIfModelReady(analyzeInformation.AnalyzeDestination, analyzeInformation.AnalyzedModel, analyzeInformation.AnalyzeToken),
	}
}

func UploadEntriesImpl(token string, model string, envPrefix string) {
	analyzeInformation.IsAnalyzing = true
	analyzeInformation.AnalyzedModel = model
	analyzeInformation.AnalyzeToken = token
	analyzeInformation.AnalyzeDestination = envPrefix

	sleepTime := time.Second * 10

	var timestampFrom int64 = 0

	for {
		timestampTo := time.Now().UnixNano() / int64(time.Millisecond)
		fmt.Printf("Getting entries from %v, to %v\n", timestampFrom, timestampTo)
		entriesArray := database.GetEntriesFromDb(timestampFrom, timestampTo)

		if len(entriesArray) > 0 {
			fmt.Printf("About to upload %v entries\n", len(entriesArray))

			body, jMarshalErr := json.Marshal(entriesArray)
			if jMarshalErr != nil {
				analyzeInformation.Reset()
				fmt.Println("Stopping analyzing")
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
				log.Println("Stopping analyzing")
				log.Fatal(postErr)
			}
			fmt.Printf("Finish uploading %v entries to %s\n", len(entriesArray), GetTrafficDumpUrl(envPrefix, model))

		} else {
			fmt.Println("Nothing to upload")
		}

		fmt.Printf("Sleeping for %v...\n", sleepTime)
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
