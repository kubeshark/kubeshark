package helm

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/kubeshark/kubeshark/config"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/repo"
)

var settings = cli.New()

type Helm struct{}

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

func (helm *Helm) Install() {
	kubeConfigPath := config.Config.KubeConfigPath()
	releaseName := "kubeshark"
	releaseNamespace := "default"
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(kube.GetConfig(kubeConfigPath, "", releaseNamespace), releaseNamespace, os.Getenv("HELM_DRIVER"), func(format string, v ...interface{}) {
		fmt.Printf(format+"\n", v)
	}); err != nil {
		panic(err)
	}

	client := action.NewInstall(actionConfig)
	client.Namespace = releaseNamespace
	client.ReleaseName = releaseName

	chartURL, err := repo.FindChartInRepoURL("https://helm.kubeshark.co", "kubeshark", "", "", "", "", getter.All(&cli.EnvSettings{}))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Chart URL: %+v\n", chartURL)

	cp, err := client.ChartPathOptions.LocateChart(chartURL, settings)
	if err != nil {
		panic(err)
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
		panic(err)
	}

	version := ""
	if registry.IsOCI(chartURL) {
		chartURL, version, err = parseOCIRef(chartURL)
		if err != nil {
			panic(errors.Wrapf(err, "could not parse OCI reference"))
		}
		dl.Options = append(dl.Options,
			getter.WithRegistryClient(m.RegistryClient),
			getter.WithTagName(version))
	}

	if _, _, err = dl.DownloadTo(chartURL, version, repoPath); err != nil {
		panic(errors.Wrapf(err, "could not download %s", chartURL))
	}

	// chartPath := "./kubeshark-40.5.tgz"
	chart, err := loader.Load(m.ChartPath)
	if err != nil {
		panic(err)
	}

	rel, err := client.Run(chart, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully installed release: ", rel.Name)
}

func (helm *Helm) Uninstall() {
	kubeConfigPath := config.Config.KubeConfigPath()
	releaseName := "kubeshark"
	releaseNamespace := "default"
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(kube.GetConfig(kubeConfigPath, "", releaseNamespace), releaseNamespace, os.Getenv("HELM_DRIVER"), func(format string, v ...interface{}) {
		fmt.Printf(format+"\n", v)
	}); err != nil {
		panic(err)
	}

	client := action.NewUninstall(actionConfig)

	resp, err := client.Run(releaseName)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s: %s\n", resp.Info, resp.Release.Name)
}
