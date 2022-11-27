package version

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/kubeshark/kubeshark/kubeshark"
	"github.com/kubeshark/kubeshark/pkg/version"

	"github.com/google/go-github/v37/github"
)

func CheckNewerVersion(versionChan chan string) {
	log.Printf("Checking for newer version...")
	start := time.Now()
	client := github.NewClient(nil)
	latestRelease, _, err := client.Repositories.GetLatestRelease(context.Background(), "kubeshark", "kubeshark")
	if err != nil {
		log.Printf("[ERROR] Failed to get latest release")
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
		log.Printf("[ERROR] Version file not found in the latest release")
		versionChan <- ""
		return
	}

	res, err := http.Get(versionFileUrl)
	if err != nil {
		log.Printf("[ERROR] Failed to get the version file %v", err)
		versionChan <- ""
		return
	}

	data, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Printf("[ERROR] Failed to read the version file -> %v", err)
		versionChan <- ""
		return
	}
	gitHubVersion := string(data)
	gitHubVersion = gitHubVersion[:len(gitHubVersion)-1]

	greater, err := version.GreaterThen(gitHubVersion, kubeshark.Ver)
	if err != nil {
		log.Printf("[ERROR] Ver version is not valid, github version %v, current version %v", gitHubVersion, kubeshark.Ver)
		versionChan <- ""
		return
	}

	log.Printf("Finished version validation, github version %v, current version %v, took %v", gitHubVersion, kubeshark.Ver, time.Since(start))

	if greater {
		var downloadCommand string
		if runtime.GOOS == "windows" {
			downloadCommand = fmt.Sprintf("curl -LO %v/kubeshark.exe", strings.Replace(*latestRelease.HTMLURL, "tag", "download", 1))
		} else {
			downloadCommand = "sh <(curl -Ls https://kubeshark.co/install)"
		}
		versionChan <- fmt.Sprintf("Update available! %v -> %v run the command: %s", kubeshark.Ver, gitHubVersion, downloadCommand)
	} else {
		versionChan <- ""
	}
}
