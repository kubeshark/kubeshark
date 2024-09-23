package connect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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
const DefaultSleep = 1 * time.Second

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
		time.Sleep(5 * DefaultSleep)
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
			var resp *http.Response
			if resp, err = utils.Post(postWorkerUrl, "application/json", bytes.NewBuffer(podMarshalled), connector.client, config.Config.License); err != nil || resp.StatusCode != http.StatusOK {
				if _, ok := err.(*url.Error); ok {
					break
				}
				log.Warn().Err(err).Msg("Failed sending the Worker pod to Hub. Retrying...")
			} else {
				log.Debug().Interface("worker-pod", pod).Msg("Reported worker pod to Hub:")
				return
			}
			time.Sleep(DefaultSleep)
		}
	}
}

type postLicenseRequest struct {
	License string `json:"license"`
}

func (connector *Connector) PostLicense(license string) {
	postLicenseUrl := fmt.Sprintf("%s/license", connector.url)

	payload := postLicenseRequest{
		License: license,
	}

	if payloadMarshalled, err := json.Marshal(payload); err != nil {
		log.Error().Err(err).Msg("Failed to marshal the payload:")
	} else {
		ok := false
		for !ok {
			var resp *http.Response
			if resp, err = utils.Post(postLicenseUrl, "application/json", bytes.NewBuffer(payloadMarshalled), connector.client, config.Config.License); err != nil || resp.StatusCode != http.StatusOK {
				if _, ok := err.(*url.Error); ok {
					break
				}
				log.Warn().Err(err).Msg("Failed sending the license to Hub. Retrying...")
			} else {
				log.Debug().Str("license", license).Msg("Reported license to Hub:")
				return
			}
			time.Sleep(DefaultSleep)
		}
	}
}

func (connector *Connector) PostPcapsMerge(out *os.File) {
	postEnvUrl := fmt.Sprintf("%s/pcaps/merge", connector.url)

	if envMarshalled, err := json.Marshal(map[string]string{"query": ""}); err != nil {
		log.Error().Err(err).Msg("Failed to marshal the env:")
	} else {
		ok := false
		for !ok {
			var resp *http.Response
			if resp, err = utils.Post(postEnvUrl, "application/json", bytes.NewBuffer(envMarshalled), connector.client, config.Config.License); err != nil || resp.StatusCode != http.StatusOK {
				if _, ok := err.(*url.Error); ok {
					break
				}
				log.Warn().Err(err).Msg("Failed exported PCAP download. Retrying...")
			} else {
				defer resp.Body.Close()

				// Check server response
				if resp.StatusCode != http.StatusOK {
					log.Error().Str("status", resp.Status).Err(err).Msg("Failed exported PCAP download.")
					return
				}

				// Writer the body to file
				_, err = io.Copy(out, resp.Body)
				if err != nil {
					log.Error().Err(err).Msg("Failed writing PCAP export:")
					return
				}
				log.Info().Str("path", out.Name()).Msg("Downloaded exported PCAP:")
				return
			}
			time.Sleep(DefaultSleep)
		}
	}
}
