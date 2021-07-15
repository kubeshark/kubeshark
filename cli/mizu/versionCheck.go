package mizu

import (
	"encoding/json"
	"fmt"
	"github.com/up9inc/mizu/shared"
	"net/http"
	"net/url"
)


func getApiVersion(port uint16) (string, error) {
	versionUrl, _ := url.Parse(fmt.Sprintf("http://localhost:%d/mizu/metadata/version", port))
	req := &http.Request{
		Method: http.MethodGet,
		URL:    versionUrl,
	}
	statusResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer statusResp.Body.Close()

	versionResponse := &shared.VersionResponse{}
	_ = json.NewDecoder(statusResp.Body).Decode(&versionResponse)

	return versionResponse.SemVer, nil
}

func CheckVersionCompatibility(port uint16) bool {
	apiSemVer, err := getApiVersion(port)
	if err != nil {
		return true
	}
	if apiSemVer == SemVer {
		return true
	}
	fmt.Printf(Red, fmt.Sprintf("cli version (%s) is not compatible with api version (%s)\n", SemVer, apiSemVer))
	return false
}