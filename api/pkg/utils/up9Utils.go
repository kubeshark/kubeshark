package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
	statusUrl, _ :=  url.Parse(fmt.Sprintf("https://trcc.%s/models/%s/status", analyzeDestination, analyzeModel))
	req := &http.Request{
		Method: http.MethodGet,
		URL:    statusUrl,
		Header: map[string][]string{
			"Content-Type":     {"application/json"},
			"Guest-Auth":       {analyzeToken},
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

