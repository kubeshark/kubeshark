package version

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/cli/pkg/version"
	"github.com/up9inc/mizu/logger"

	"github.com/google/go-github/v37/github"
)

func CheckNewerVersion(versionChan chan string) {
	logger.Log.Debugf("Checking for newer version...")
	start := time.Now()
	client := github.NewClient(nil)
	latestRelease, _, err := client.Repositories.GetLatestRelease(context.Background(), "up9inc", "mizu")
	if err != nil {
		logger.Log.Debugf("[ERROR] Failed to get latest release")
		versionChan <- ""
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
		logger.Log.Debugf("[ERROR] Version file not found in the latest release")
		versionChan <- ""
		return
	}

	res, err := http.Get(versionFileUrl)
	if err != nil {
		logger.Log.Debugf("[ERROR] Failed to get the version file %v", err)
		versionChan <- ""
		return
	}

	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		logger.Log.Debugf("[ERROR] Failed to read the version file -> %v", err)
		versionChan <- ""
		return
	}
	gitHubVersion := string(data)
	gitHubVersion = gitHubVersion[:len(gitHubVersion)-1]

	greater, err := version.GreaterThen(gitHubVersion, mizu.Ver)
	if err != nil {
		logger.Log.Debugf("[ERROR] Ver version is not valid, github version %v, current version %v", gitHubVersion, mizu.Ver)
		versionChan <- ""
		return
	}

	logger.Log.Debugf("Finished version validation, github version %v, current version %v, took %v", gitHubVersion, mizu.Ver, time.Since(start))

	if greater {
		var downloadMessage string
		if runtime.GOOS == "windows" {
			downloadMessage = fmt.Sprintf("curl -LO %v/mizu.exe", strings.Replace(*latestRelease.HTMLURL, "tag", "download", 1))
		} else {
			downloadMessage = fmt.Sprintf("curl -Lo mizu %v/mizu_%s_%s && chmod 755 mizu", strings.Replace(*latestRelease.HTMLURL, "tag", "download", 1), runtime.GOOS, runtime.GOARCH)
		}
		versionChan <- fmt.Sprintf("Update available! %v -> %v (%s)", mizu.Ver, gitHubVersion, downloadMessage)
	} else {
		versionChan <- ""
	}
}
