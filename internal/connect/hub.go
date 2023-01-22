package connect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/utils"

	"github.com/rs/zerolog/log"
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
		retries: retries,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (connector *Connector) TestConnection(path string) error {
	retriesLeft := connector.retries
	for retriesLeft > 0 {
		if isReachable, err := connector.isReachable(path); err != nil || !isReachable {
			log.Debug().Str("url", connector.url).Err(err).Msg("Not ready yet!")
		} else {
			log.Debug().Str("url", connector.url).Msg("Connection test passed successfully.")
			break
		}
		retriesLeft -= 1
		time.Sleep(time.Second)
	}

	if retriesLeft == 0 {
		return fmt.Errorf("Couldn't reach the URL: %s after %d retries!", connector.url, connector.retries)
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
				connector.PostStorageLimitToHub(config.Config.Tap.StorageLimitBytes())
			}
			time.Sleep(time.Second)
		}
	}
}

type postStorageLimit struct {
	Limit int64 `json:"limit"`
}

func (connector *Connector) PostStorageLimitToHub(limit int64) {
	payload := &postStorageLimit{
		Limit: limit,
	}
	postStorageLimitUrl := fmt.Sprintf("%s/pcaps/set-storage-limit", connector.url)

	if payloadMarshalled, err := json.Marshal(payload); err != nil {
		log.Error().Err(err).Msg("Failed to marshal the storage limit:")
	} else {
		ok := false
		for !ok {
			if _, err = utils.Post(postStorageLimitUrl, "application/json", bytes.NewBuffer(payloadMarshalled), connector.client); err != nil {
				if _, ok := err.(*url.Error); ok {
					break
				}
				log.Debug().Err(err).Msg("Failed sending the storage limit to Hub:")
			} else {
				ok = true
				log.Debug().Int("limit", int(limit)).Msg("Reported storage limit to Hub:")
			}
			time.Sleep(time.Second)
		}
	}
}

type postRegexRequest struct {
	Regex      string   `json:"regex"`
	Namespaces []string `json:"namespaces"`
}

func (connector *Connector) PostRegexToHub(regex string, namespaces []string) {
	postRegexUrl := fmt.Sprintf("%s/pods/regex", connector.url)

	payload := postRegexRequest{
		Regex:      regex,
		Namespaces: namespaces,
	}

	if payloadMarshalled, err := json.Marshal(payload); err != nil {
		log.Error().Err(err).Msg("Failed to marshal the payload:")
	} else {
		ok := false
		for !ok {
			if _, err = utils.Post(postRegexUrl, "application/json", bytes.NewBuffer(payloadMarshalled), connector.client); err != nil {
				if _, ok := err.(*url.Error); ok {
					break
				}
				log.Debug().Err(err).Msg("Failed sending the payload to Hub:")
			} else {
				ok = true
				log.Debug().Str("regex", regex).Strs("namespaces", namespaces).Msg("Reported payload to Hub:")
			}
			time.Sleep(time.Second)
		}
	}
}
