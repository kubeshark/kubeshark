package mizu

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/google/go-github/v37/github"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/semver"
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

	Log.Infof(uiUtils.Red, fmt.Sprintf("cli version (%s) is not compatible with api version (%s)", SemVer, apiSemVer))
	return false, nil
}

func CheckNewerVersion() {
	Log.Debugf("Checking for newer version...")
	start := time.Now()
	client := github.NewClient(nil)
	latestRelease, _, err := client.Repositories.GetLatestRelease(context.Background(), "up9inc", "mizu")
	if err != nil {
		Log.Debugf("Failed to get latest release")
		return
	}

	versionFileUrl := ""
	for _, asset := range latestRelease.Assets {
		if *asset.Name == "version.txt" {
			versionFileUrl = *asset.BrowserDownloadURL
			break
		}
	}
	if versionFileUrl == "" {
		Log.Debugf("Version file not found in the latest release")
		return
	}

	res, err := http.Get(versionFileUrl)
	if err != nil {
		Log.Debugf("http.Get version asset -> %v", err)
		return
	}

	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		Log.Debugf("ioutil.ReadAll -> %v", err)
		return
	}
	gitHubVersion := string(data)
	gitHubVersion = gitHubVersion[:len(gitHubVersion)-1]
	Log.Debugf("Finished version validation, took %v", time.Since(start))
	if SemVer < gitHubVersion {
		Log.Infof(uiUtils.Yellow, fmt.Sprintf("Update available! %v -> %v (%v)", SemVer, gitHubVersion, *latestRelease.HTMLURL))
	}
}
