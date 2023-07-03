package helm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/misc"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
)

const ENV_HELM_DRIVER = "HELM_DRIVER"

var settings = cli.New()

type Helm struct {
	repo             string
	releaseName      string
	releaseNamespace string
}

func NewHelm(repo string, releaseName string, releaseNamespace string) *Helm {
	return &Helm{
		repo:             repo,
		releaseName:      releaseName,
		releaseNamespace: releaseNamespace,
	}
}

func parseOCIRef(chartRef string) (string, string, error) {
	refTagRegexp := regexp.MustCompile(`^(oci://[^:]+(:[0-9]{1,5})?[^:]+):(.*)$`)
	caps := refTagRegexp.FindStringSubmatch(chartRef)
	if len(caps) != 4 {
		return "", "", errors.Errorf("improperly formatted oci chart reference: %s", chartRef)
	}
	chartRef = caps[1]
	tag := caps[3]

	return chartRef, tag, nil
}

func (h *Helm) Install() (rel *release.Release, err error) {
	kubeConfigPath := config.Config.KubeConfigPath()
	actionConfig := new(action.Configuration)
	if err = actionConfig.Init(kube.GetConfig(kubeConfigPath, "", h.releaseNamespace), h.releaseNamespace, os.Getenv(ENV_HELM_DRIVER), func(format string, v ...interface{}) {
		log.Info().Msgf(format, v...)
	}); err != nil {
		return
	}

	client := action.NewInstall(actionConfig)
	client.Namespace = h.releaseNamespace
	client.ReleaseName = h.releaseName

	chartPath := os.Getenv(fmt.Sprintf("%s_HELM_CHART_PATH", strings.ToUpper(misc.Program)))
	if chartPath == "" {
		var chartURL string
		chartURL, err = repo.FindChartInRepoURL(h.repo, h.releaseName, "", "", "", "", getter.All(&cli.EnvSettings{}))
		if err != nil {
			return
		}

		var cp string
		cp, err = client.ChartPathOptions.LocateChart(chartURL, settings)
		if err != nil {
			return
		}

		m := &downloader.Manager{
			Out:              os.Stdout,
			ChartPath:        cp,
			Keyring:          client.ChartPathOptions.Keyring,
			SkipUpdate:       false,
			Getters:          getter.All(settings),
			RepositoryConfig: settings.RepositoryConfig,
			RepositoryCache:  settings.RepositoryCache,
			Debug:            settings.Debug,
		}

		dl := downloader.ChartDownloader{
			Out:              m.Out,
			Verify:           m.Verify,
			Keyring:          m.Keyring,
			RepositoryConfig: m.RepositoryConfig,
			RepositoryCache:  m.RepositoryCache,
			RegistryClient:   m.RegistryClient,
			Getters:          m.Getters,
			Options: []getter.Option{
				getter.WithInsecureSkipVerifyTLS(false),
			},
		}

		repoPath := filepath.Dir(m.ChartPath)
		err = os.MkdirAll(repoPath, os.ModePerm)
		if err != nil {
			return
		}

		version := ""
		if registry.IsOCI(chartURL) {
			chartURL, version, err = parseOCIRef(chartURL)
			if err != nil {
				return
			}
			dl.Options = append(dl.Options,
				getter.WithRegistryClient(m.RegistryClient),
				getter.WithTagName(version))
		}

		log.Info().
			Str("url", chartURL).
			Str("repo-path", repoPath).
			Msg("Downloading Helm chart:")

		if _, _, err = dl.DownloadTo(chartURL, version, repoPath); err != nil {
			return
		}

		chartPath = m.ChartPath
	}
	var chart *chart.Chart
	chart, err = loader.Load(chartPath)
	if err != nil {
		return
	}

	log.Info().
		Str("release", chart.Metadata.Name).
		Str("version", chart.Metadata.Version).
		Strs("source", chart.Metadata.Sources).
		Str("kube-version", chart.Metadata.KubeVersion).
		Msg("Installing using Helm:")

	var configMarshalled []byte
	configMarshalled, err = json.Marshal(config.Config)
	if err != nil {
		return
	}

	var configUnmarshalled map[string]interface{}
	err = json.Unmarshal(configMarshalled, &configUnmarshalled)
	if err != nil {
		return
	}

	rel, err = client.Run(chart, configUnmarshalled)
	if err != nil {
		return
	}

	return
}

func (h *Helm) Uninstall() (resp *release.UninstallReleaseResponse, err error) {
	kubeConfigPath := config.Config.KubeConfigPath()
	actionConfig := new(action.Configuration)
	if err = actionConfig.Init(kube.GetConfig(kubeConfigPath, "", h.releaseNamespace), h.releaseNamespace, os.Getenv(ENV_HELM_DRIVER), func(format string, v ...interface{}) {
		log.Info().Msgf(format, v...)
	}); err != nil {
		return
	}

	client := action.NewUninstall(actionConfig)

	resp, err = client.Run(h.releaseName)
	if err != nil {
		return
	}

	return
}
