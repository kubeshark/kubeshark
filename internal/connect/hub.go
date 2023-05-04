package connect

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/kubeshark/kubeshark/misc"
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
			if resp, err = utils.Post(postWorkerUrl, "application/json", bytes.NewBuffer(podMarshalled), connector.client); err != nil || resp.StatusCode != http.StatusOK {
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
		log.Error().Err(err).Msg("Failed to marshal the pod regex:")
	} else {
		ok := false
		for !ok {
			var resp *http.Response
			if resp, err = utils.Post(postRegexUrl, "application/json", bytes.NewBuffer(payloadMarshalled), connector.client); err != nil || resp.StatusCode != http.StatusOK {
				if _, ok := err.(*url.Error); ok {
					break
				}
				log.Warn().Err(err).Msg("Failed sending the pod regex to Hub. Retrying...")
			} else {
				log.Debug().Str("regex", regex).Strs("namespaces", namespaces).Msg("Reported pod regex to Hub:")
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
			if resp, err = utils.Post(postLicenseUrl, "application/json", bytes.NewBuffer(payloadMarshalled), connector.client); err != nil || resp.StatusCode != http.StatusOK {
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

func (connector *Connector) PostLicenseSingle(license string) {
	postLicenseUrl := fmt.Sprintf("%s/license", connector.url)

	payload := postLicenseRequest{
		License: license,
	}

	if payloadMarshalled, err := json.Marshal(payload); err != nil {
		log.Error().Err(err).Msg("Failed to marshal the payload:")
	} else {
		var resp *http.Response
		if resp, err = utils.Post(postLicenseUrl, "application/json", bytes.NewBuffer(payloadMarshalled), connector.client); err != nil || resp.StatusCode != http.StatusOK {
			log.Warn().Err(err).Msg("Failed sending the license to Hub.")
		} else {
			log.Debug().Str("license", license).Msg("Reported license to Hub:")
			return
		}
	}
}

func (connector *Connector) PostEnv(env map[string]interface{}) {
	if len(env) == 0 {
		return
	}

	postEnvUrl := fmt.Sprintf("%s/scripts/env", connector.url)

	if envMarshalled, err := json.Marshal(env); err != nil {
		log.Error().Err(err).Msg("Failed to marshal the env:")
	} else {
		ok := false
		for !ok {
			var resp *http.Response
			if resp, err = utils.Post(postEnvUrl, "application/json", bytes.NewBuffer(envMarshalled), connector.client); err != nil || resp.StatusCode != http.StatusOK {
				if _, ok := err.(*url.Error); ok {
					break
				}
				log.Warn().Err(err).Msg("Failed sending the scripting environment variables to Hub. Retrying...")
			} else {
				log.Debug().Interface("env", env).Msg("Reported scripting environment variables to Hub:")
				return
			}
			time.Sleep(DefaultSleep)
		}
	}
}

func (connector *Connector) PostScript(script *misc.Script) (index int64, err error) {
	postScriptUrl := fmt.Sprintf("%s/scripts", connector.url)

	var scriptMarshalled []byte
	if scriptMarshalled, err = json.Marshal(script); err != nil {
		log.Error().Err(err).Msg("Failed to marshal the script:")
	} else {
		ok := false
		for !ok {
			var resp *http.Response
			if resp, err = utils.Post(postScriptUrl, "application/json", bytes.NewBuffer(scriptMarshalled), connector.client); err != nil || resp.StatusCode != http.StatusOK {
				if _, ok := err.(*url.Error); ok {
					break
				}
				log.Warn().Err(err).Msg("Failed creating script Hub:")
			} else {

				var j map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&j)
				if err != nil {
					return
				}

				val, ok := j["index"]
				if !ok {
					err = errors.New("Response does not contain `key` field!")
					return
				}

				index = int64(val.(float64))

				log.Debug().Int("index", int(index)).Interface("script", script).Msg("Created script on Hub:")
				return
			}
			time.Sleep(DefaultSleep)
		}
	}

	return
}

func (connector *Connector) PutScript(script *misc.Script, index int64) (err error) {
	putScriptUrl := fmt.Sprintf("%s/scripts/%d", connector.url, index)

	var scriptMarshalled []byte
	if scriptMarshalled, err = json.Marshal(script); err != nil {
		log.Error().Err(err).Msg("Failed to marshal the script:")
	} else {
		ok := false
		for !ok {
			client := &http.Client{}

			var req *http.Request
			req, err = http.NewRequest(http.MethodPut, putScriptUrl, bytes.NewBuffer(scriptMarshalled))
			if err != nil {
				log.Error().Err(err).Send()
				return
			}
			req.Header.Set("Content-Type", "application/json")

			var resp *http.Response
			resp, err = client.Do(req)
			if err != nil {
				log.Error().Err(err).Send()
				return
			}

			if resp.StatusCode != http.StatusOK {
				if _, ok := err.(*url.Error); ok {
					break
				}
				log.Warn().Err(err).Msg("Failed updating script on Hub:")
			} else {
				log.Debug().Int("index", int(index)).Interface("script", script).Msg("Updated script on Hub:")
				return
			}
			time.Sleep(DefaultSleep)
		}
	}

	return
}

func (connector *Connector) DeleteScript(index int64) (err error) {
	deleteScriptUrl := fmt.Sprintf("%s/scripts/%d", connector.url, index)

	ok := false
	for !ok {
		client := &http.Client{}

		var req *http.Request
		req, err = http.NewRequest(http.MethodDelete, deleteScriptUrl, nil)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}
		req.Header.Set("Content-Type", "application/json")

		var resp *http.Response
		resp, err = client.Do(req)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}

		if resp.StatusCode != http.StatusOK {
			if _, ok := err.(*url.Error); ok {
				break
			}
			log.Warn().Err(err).Msg("Failed deleting script on Hub:")
		} else {
			log.Debug().Int("index", int(index)).Msg("Deleted script on Hub:")
			return
		}
		time.Sleep(DefaultSleep)
	}

	return
}

func (connector *Connector) PostScriptDone() {
	postScripDonetUrl := fmt.Sprintf("%s/scripts/done", connector.url)

	ok := false
	var err error
	for !ok {
		var resp *http.Response
		if resp, err = utils.Post(postScripDonetUrl, "application/json", nil, connector.client); err != nil || resp.StatusCode != http.StatusOK {
			if _, ok := err.(*url.Error); ok {
				break
			}
			log.Warn().Err(err).Msg("Failed sending the POST scripts done to Hub. Retrying...")
		} else {
			log.Debug().Msg("Reported POST scripts done to Hub.")
			return
		}
		time.Sleep(DefaultSleep)
	}
}
