package connect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/kubeshark/kubeshark/utils"
	"github.com/kubeshark/worker/models"

	"github.com/kubeshark/kubeshark/config"
	core "k8s.io/api/core/v1"
)

type Connector struct {
	url     string
	retries int
	client  *http.Client
}

const DefaultRetries = 3
const DefaultTimeout = 2 * time.Second

func NewConnector(url string, retries int, timeout time.Duration) *Connector {
	return &Connector{
		url:     url,
		retries: config.GetIntEnvConfig(config.ApiServerRetries, retries),
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (connector *Connector) TestConnection(path string) error {
	retriesLeft := connector.retries
	for retriesLeft > 0 {
		if isReachable, err := connector.isReachable(path); err != nil || !isReachable {
			log.Printf("api server not ready yet %v", err)
		} else {
			log.Printf("connection test to api server passed successfully")
			break
		}
		retriesLeft -= 1
		time.Sleep(time.Second)
	}

	if retriesLeft == 0 {
		return fmt.Errorf("couldn't reach the api server after %v retries", connector.retries)
	}
	return nil
}

func (connector *Connector) isReachable(path string) (bool, error) {
	targetUrl := fmt.Sprintf("%s%s", connector.url, path)
	if _, err := utils.Get(targetUrl, connector.client); err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func (connector *Connector) ReportTapperStatus(tapperStatus models.TapperStatus) error {
	tapperStatusUrl := fmt.Sprintf("%s/status/tapperStatus", connector.url)

	if jsonValue, err := json.Marshal(tapperStatus); err != nil {
		return fmt.Errorf("failed Marshal the tapper status %w", err)
	} else {
		if _, err := utils.Post(tapperStatusUrl, "application/json", bytes.NewBuffer(jsonValue), connector.client); err != nil {
			return fmt.Errorf("failed sending to API server the tapped pods %w", err)
		} else {
			log.Printf("Reported to server API about tapper status: %v", tapperStatus)
			return nil
		}
	}
}

func (connector *Connector) ReportTappedPods(pods []core.Pod) error {
	tappedPodsUrl := fmt.Sprintf("%s/status/tappedPods", connector.url)

	if jsonValue, err := json.Marshal(pods); err != nil {
		return fmt.Errorf("failed Marshal the tapped pods %w", err)
	} else {
		if _, err := utils.Post(tappedPodsUrl, "application/json", bytes.NewBuffer(jsonValue), connector.client); err != nil {
			return fmt.Errorf("failed sending to API server the tapped pods %w", err)
		} else {
			log.Printf("Reported to server API about %d taped pods successfully", len(pods))
			return nil
		}
	}
}
