package connect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kubeshark/kubeshark/utils"

	"github.com/kubeshark/kubeshark/config"
	"github.com/rs/zerolog/log"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
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
		retries: config.GetIntEnvConfig(config.HubRetries, retries),
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (connector *Connector) TestConnection(path string) error {
	retriesLeft := connector.retries
	for retriesLeft > 0 {
		if isReachable, err := connector.isReachable(path); err != nil || !isReachable {
			log.Debug().Err(err).Msg("Hub is not ready yet!")
		} else {
			log.Debug().Msg("Connection test to Hub passed successfully.")
			break
		}
		retriesLeft -= 1
		time.Sleep(time.Second)
	}

	if retriesLeft == 0 {
		return fmt.Errorf("Couldn't reach the Hub after %d retries!", connector.retries)
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

func (connector *Connector) PostWorkerPodToHub(pod *v1.Pod) error {
	setWorkerUrl := fmt.Sprintf("%s/pods/set-worker", connector.url)

	if jsonValue, err := json.Marshal(pod); err != nil {
		return fmt.Errorf("Failed to marshal the Worker pod: %w", err)
	} else {
		if _, err := utils.Post(setWorkerUrl, "application/json", bytes.NewBuffer(jsonValue), connector.client); err != nil {
			return fmt.Errorf("Failed sending the Worker pod to Hub: %w", err)
		} else {
			log.Debug().Interface("worker-pod", pod).Msg("Reported to Hub about Worker status:")
			return nil
		}
	}
}

func (connector *Connector) ReportTargettedPods(pods []core.Pod) error {
	targettedPodsUrl := fmt.Sprintf("%s/status/targettedPods", connector.url)

	if jsonValue, err := json.Marshal(pods); err != nil {
		return fmt.Errorf("Failed Marshal the targetted pods %w", err)
	} else {
		if _, err := utils.Post(targettedPodsUrl, "application/json", bytes.NewBuffer(jsonValue), connector.client); err != nil {
			return fmt.Errorf("Failed sending to Hub the targetted pods %w", err)
		} else {
			log.Debug().Int("pod-count", len(pods)).Msg("Reported to Hub about targetted pod count:")
			return nil
		}
	}
}
