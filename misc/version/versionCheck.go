package version

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/kubeshark/kubeshark/misc"
	"github.com/kubeshark/kubeshark/pkg/version"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/rs/zerolog/log"

	"github.com/google/go-github/v37/github"
)

func CheckNewerVersion() {
	log.Info().Msg("Checking for a newer version...")
	start := time.Now()
	client := github.NewClient(nil)
	latestRelease, _, err := client.Repositories.GetLatestRelease(context.Background(), misc.Program, misc.Program)
	if err != nil {
		log.Error().Msg("Failed to get latest release.")
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
		log.Error().Msg("Version file not found in the latest release.")
		return
	}

	res, err := http.Get(versionFileUrl)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get the version file.")
		return
	}

	data, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Error().Err(err).Msg("Failed to read the version file.")
		return
	}
	gitHubVersion := string(data)
	gitHubVersion = gitHubVersion[:len(gitHubVersion)-1]

	greater, err := version.GreaterThen(gitHubVersion, misc.Ver)
	if err != nil {
		log.Error().
			Str("upstream-version", gitHubVersion).
			Str("local-version", misc.Ver).
			Msg("Version is invalid!")
		return
	}

	log.Debug().
		Str("upstream-version", gitHubVersion).
		Str("local-version", misc.Ver).
		Dur("elapsed-time", time.Since(start)).
		Msg("Finished version validation.")

	if greater {
		var downloadCommand string
		if runtime.GOOS == "windows" {
			downloadCommand = fmt.Sprintf("curl -LO %v/%s.exe", strings.Replace(*latestRelease.HTMLURL, "tag", "download", 1), misc.Program)
		} else {
			downloadCommand = fmt.Sprintf("sh <(curl -Ls %s/install)", misc.Website)
		}
		msg := fmt.Sprintf("Update available! %v -> %v run:", misc.Ver, gitHubVersion)
		log.Warn().Str("command", downloadCommand).Msg(fmt.Sprintf(utils.Yellow, msg))
	}
}
