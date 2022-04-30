package apiserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/up9inc/mizu/cli/utils"

	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/shared"
	core "k8s.io/api/core/v1"
)

type Provider struct {
	url     string
	retries int
	client  *http.Client
}

const DefaultRetries = 3
const DefaultTimeout = 2 * time.Second

func NewProvider(url string, retries int, timeout time.Duration) *Provider {
	return &Provider{
		url:     url,
		retries: config.GetIntEnvConfig(config.ApiServerRetries, retries),
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (provider *Provider) TestConnection() error {
	retriesLeft := provider.retries
	for retriesLeft > 0 {
		if isReachable, err := provider.isReachable(); err != nil || !isReachable {
			logger.Log.Debugf("api server not ready yet %v", err)
		} else {
			logger.Log.Debugf("connection test to api server passed successfully")
			break
		}
		retriesLeft -= 1
		time.Sleep(time.Second)
	}

	if retriesLeft == 0 {
		return fmt.Errorf("couldn't reach the api server after %v retries", provider.retries)
	}
	return nil
}

func (provider *Provider) isReachable() (bool, error) {
	echoUrl := fmt.Sprintf("%s/echo", provider.url)
	if _, err := utils.Get(echoUrl, provider.client); err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func (provider *Provider) ReportTapperStatus(tapperStatus shared.TapperStatus) error {
	tapperStatusUrl := fmt.Sprintf("%s/status/tapperStatus", provider.url)

	if jsonValue, err := json.Marshal(tapperStatus); err != nil {
		return fmt.Errorf("failed Marshal the tapper status %w", err)
	} else {
		if _, err := utils.Post(tapperStatusUrl, "application/json", bytes.NewBuffer(jsonValue), provider.client); err != nil {
			return fmt.Errorf("failed sending to API server the tapped pods %w", err)
		} else {
			logger.Log.Debugf("Reported to server API about tapper status: %v", tapperStatus)
			return nil
		}
	}
}

func (provider *Provider) ReportTappedPods(pods []core.Pod) error {
	tappedPodsUrl := fmt.Sprintf("%s/status/tappedPods", provider.url)

	if jsonValue, err := json.Marshal(pods); err != nil {
		return fmt.Errorf("failed Marshal the tapped pods %w", err)
	} else {
		if _, err := utils.Post(tappedPodsUrl, "application/json", bytes.NewBuffer(jsonValue), provider.client); err != nil {
			return fmt.Errorf("failed sending to API server the tapped pods %w", err)
		} else {
			logger.Log.Debugf("Reported to server API about %d taped pods successfully", len(pods))
			return nil
		}
	}
}

func (provider *Provider) GetGeneralStats() (map[string]interface{}, error) {
	generalStatsUrl := fmt.Sprintf("%s/status/general", provider.url)

	response, requestErr := utils.Get(generalStatsUrl, provider.client)
	if requestErr != nil {
		return nil, fmt.Errorf("failed to get general stats for telemetry, err: %w", requestErr)
	}

	defer response.Body.Close()

	data, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read general stats for telemetry, err: %v", readErr)
	}

	var generalStats map[string]interface{}
	if parseErr := json.Unmarshal(data, &generalStats); parseErr != nil {
		return nil, fmt.Errorf("failed to parse general stats for telemetry, err: %v", parseErr)
	}
	return generalStats, nil
}

func (provider *Provider) GetVersion() (string, error) {
	versionUrl, _ := url.Parse(fmt.Sprintf("%s/metadata/version", provider.url))
	req := &http.Request{
		Method: http.MethodGet,
		URL:    versionUrl,
	}
	statusResp, err := utils.Do(req, provider.client)
	if err != nil {
		return "", err
	}
	defer statusResp.Body.Close()

	versionResponse := &shared.VersionResponse{}
	if err := json.NewDecoder(statusResp.Body).Decode(&versionResponse); err != nil {
		return "", err
	}

	return versionResponse.Ver, nil
}
