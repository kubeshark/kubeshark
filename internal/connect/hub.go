package connect

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/misc"
	"github.com/kubeshark/kubeshark/utils"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

const (
	DefaultRetries = 3
	DefaultTimeout = 2 * time.Second
	DefaultSleep   = 1 * time.Second
)

type Connector struct {
	baseURL string
	retries int
	client  *http.Client
}

func NewConnector(baseURL string, retries int, timeout time.Duration) *Connector {
	return &Connector{
		baseURL: baseURL,
		retries: retries,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Connector) TestConnection(path string) error {
	for i := 0; i < c.retries; i++ {
		if reachable, err := c.isReachable(path); err != nil || !reachable {
			log.Debug().Str("url", c.baseURL).Err(err).Msg("Not ready yet!")
		} else {
			log.Debug().Str("url", c.baseURL).Msg("Connection test passed successfully.")
			return nil
		}
		time.Sleep(DefaultSleep * 5)
	}
	return fmt.Errorf("Couldn't reach the URL: %s after %d retries!", c.baseURL, c.retries)
}

func (c *Connector) isReachable(path string) (bool, error) {
	targetURL := fmt.Sprintf("%s%s", c.baseURL, path)
	_, err := utils.Get(targetURL, c.client)
	return err == nil, err
}

func (c *Connector) postJSON(url string, payload interface{}) (*http.Response, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal payload")
		return nil, err
	}
	return utils.Post(url, "application/json", bytes.NewBuffer(data), c.client, config.Config.License)
}

func (c *Connector) retryPostJSON(url string, payload interface{}) (*http.Response, error) {
	for i := 0; i < c.retries; i++ {
		resp, err := c.postJSON(url, payload)
		if err == nil && resp.StatusCode == http.StatusOK {
			return resp, nil
		}
		log.Warn().Err(err).Msg("Retrying...")
		time.Sleep(DefaultSleep)
	}
	return nil, fmt.Errorf("failed to POST to %s after %d retries", url, c.retries)
}

func (c *Connector) PostWorkerPodToHub(pod *v1.Pod) {
	url := fmt.Sprintf("%s/pods/worker", c.baseURL)
	resp, err := c.retryPostJSON(url, pod)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send Worker pod to Hub")
		return
	}
	log.Debug().Interface("worker-pod", pod).Msg("Reported worker pod to Hub")
	defer resp.Body.Close()
}

type postLicenseRequest struct {
	License string `json:"license"`
}

func (c *Connector) PostLicense(license string) {
	url := fmt.Sprintf("%s/license", c.baseURL)
	payload := postLicenseRequest{License: license}
	resp, err := c.retryPostJSON(url, payload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send license to Hub")
		return
	}
	log.Debug().Str("license", license).Msg("Reported license to Hub")
	defer resp.Body.Close()
}

type postScriptRequest struct {
	Title string `json:"title"`
	Code  string `json:"code"`
}

func (c *Connector) PostScript(script *misc.Script) (int64, error) {
	url := fmt.Sprintf("%s/scripts", c.baseURL)
	payload := postScriptRequest{
		Title: script.Title,
		Code:  script.Code,
	}
	resp, err := c.retryPostJSON(url, payload)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	index, ok := result["index"].(float64)
	if !ok {
		return 0, errors.New("response does not contain 'index' field")
	}
	log.Debug().Int("index", int(index)).Interface("script", script).Msg("Created script on Hub")
	return int64(index), nil
}

func (c *Connector) PutScript(script *misc.Script, index int64) error {
	url := fmt.Sprintf("%s/scripts/%d", c.baseURL, index)
	client := &http.Client{}
	data, err := json.Marshal(script)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal script")
		return err
	}

	for i := 0; i < c.retries; i++ {
		req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(data))
		if err != nil {
			log.Error().Err(err).Msg("Failed to create PUT request")
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("License-Key", config.Config.License)

		resp, err := client.Do(req)
		if err == nil && resp.StatusCod
