package apiserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kubeshark/kubeshark/cli/utils"

	"github.com/kubeshark/kubeshark/cli/config"
	"github.com/kubeshark/kubeshark/logger"
	"github.com/kubeshark/kubeshark/shared"
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

func (provider *Provider) TestConnection(path string) error {
	retriesLeft := provider.retries
	for retriesLeft > 0 {
		if isReachable, err := provider.isReachable(path); err != nil || !isReachable {
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

func (provider *Provider) isReachable(path string) (bool, error) {
	targetUrl := fmt.Sprintf("%s%s", provider.url, path)
	if _, err := utils.Get(targetUrl, provider.client); err != nil {
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
