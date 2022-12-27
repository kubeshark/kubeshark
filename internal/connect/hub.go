package connect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

func (connector *Connector) PostWorkerPodToHub(pod *v1.Pod) {
	postWorkerUrl := fmt.Sprintf("%s/pods/worker", connector.url)

	if podMarshalled, err := json.Marshal(pod); err != nil {
		log.Error().Err(err).Msg("Failed to marshal the Worker pod:")
	} else {
		ok := false
		for !ok {
			if _, err = utils.Post(postWorkerUrl, "application/json", bytes.NewBuffer(podMarshalled), connector.client); err != nil {
				if _, ok := err.(*url.Error); ok {
					break
				}
				log.Debug().Err(err).Msg("Failed sending the Worker pod to Hub:")
			} else {
				ok = true
				log.Debug().Interface("worker-pod", pod).Msg("Reported worker pod to Hub:")
			}
			time.Sleep(time.Second)
		}
	}
}

func (connector *Connector) PostTargettedPodsToHub(pods []core.Pod) {
	postTargettedUrl := fmt.Sprintf("%s/pods/targetted", connector.url)

	if podsMarshalled, err := json.Marshal(pods); err != nil {
		log.Error().Err(err).Msg("Failed to marshal the targetted pods:")
	} else {
		ok := false
		for !ok {
			if _, err = utils.Post(postTargettedUrl, "application/json", bytes.NewBuffer(podsMarshalled), connector.client); err != nil {
				if _, ok := err.(*url.Error); ok {
					break
				}
				log.Debug().Err(err).Msg("Failed sending the targetted pods to Hub:")
			} else {
				ok = true
				log.Debug().Int("pod-count", len(pods)).Msg("Reported targetted pods to Hub:")
			}
			time.Sleep(time.Second)
		}
	}
}
