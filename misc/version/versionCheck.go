package version

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/kubeshark/kubeshark/misc"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/rs/zerolog/log"

	"github.com/google/go-github/v37/github"
)

func CheckNewerVersion() {
	if os.Getenv(fmt.Sprintf("%s_DISABLE_VERSION_CHECK", strings.ToUpper(misc.Program))) != "" {
		return
	}

	log.Info().Msg("Checking for a newer version...")
	start := time.Now()
	client := github.NewClient(nil)
	latestRelease, _, err := client.Repositories.GetLatestRelease(context.Background(), misc.Program, misc.Program)
	if err != nil {
		log.Error().Msg("Failed to get the latest release.")
		return
	}

	latestVersion := *latestRelease.TagName

	log.Debug().
		Str("upstream-version", latestVersion).
		Str("local-version", misc.Ver).
		Dur("elapsed-time", time.Since(start)).
		Msg("Fetched the latest release:")

	if misc.Ver != latestVersion {
		var downloadCommand string
		if runtime.GOOS == "windows" {
			downloadCommand = fmt.Sprintf("curl -LO %v/%s.exe", strings.Replace(*latestRelease.HTMLURL, "tag", "download", 1), misc.Program)
		} else {
			downloadCommand = fmt.Sprintf("sh <(curl -Ls %s/install)", misc.Website)
		}
		msg := fmt.Sprintf("There is a new release! %v -> %v Please upgrade to the latest release, as new releases are not always backward compatible. Run:", misc.Ver, latestVersion)
		log.Warn().Str("command", downloadCommand).Msg(fmt.Sprintf(utils.Yellow, msg))
	}
}
