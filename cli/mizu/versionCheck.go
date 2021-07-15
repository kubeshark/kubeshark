package mizu

import (
	"encoding/json"
	"fmt"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/semver"
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
	if err := json.NewDecoder(statusResp.Body).Decode(&versionResponse); err != nil {
		return "", err
	}

	return versionResponse.SemVer, nil
}

func CheckVersionCompatibility(port uint16) (bool, error) {
	apiSemVer, err := getApiVersion(port)
	if err != nil {
		return false, err
	}

	if semver.SemVersion(apiSemVer).Major() == semver.SemVersion(SemVer).Major() &&
		semver.SemVersion(apiSemVer).Minor() == semver.SemVersion(SemVer).Minor() {
		return true, nil
	}

	fmt.Printf(Red, fmt.Sprintf("cli version (%s) is not compatible with api version (%s)\n", SemVer, apiSemVer))
	return false, nil
}