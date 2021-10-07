package apiserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/logger"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/shared"
	"io/ioutil"
	core "k8s.io/api/core/v1"
	"net/http"
	"net/url"
	"time"
)

type apiServerProvider struct {
	url     string
	isReady bool
	retries int
}

var Provider = apiServerProvider{retries: config.GetIntEnvConfig(config.ApiServerRetries, 20)}

func (provider *apiServerProvider) InitAndTestConnection(url string) error {
	healthUrl := fmt.Sprintf("%s/", url)
	retriesLeft := provider.retries
	for retriesLeft > 0 {
		if response, err := http.Get(healthUrl); err != nil {
			logger.Log.Debugf("[ERROR] failed connecting to api server %v", err)
		} else if response.StatusCode != 200 {
			responseBody := ""
			data, readErr := ioutil.ReadAll(response.Body)
			if readErr == nil {
				responseBody = string(data)
			}

			logger.Log.Debugf("can't connect to api server yet, response status code: %v, body: %v", response.StatusCode, responseBody)

			response.Body.Close()
		} else {
			logger.Log.Debugf("connection test to api server passed successfully")
			break
		}
		retriesLeft -= 1
		time.Sleep(time.Second)
	}

	if retriesLeft == 0 {
		provider.isReady = false
		return fmt.Errorf("couldn't reach the api server after %v retries", provider.retries)
	}
	provider.url = url
	provider.isReady = true
	return nil
}

func (provider *apiServerProvider) ReportTappedPods(pods []core.Pod) error {
	if !provider.isReady {
		return fmt.Errorf("trying to reach api server when not initialized yet")
	}
	tappedPodsUrl := fmt.Sprintf("%s/status/tappedPods", provider.url)

	podInfos := make([]shared.PodInfo, 0)
	for _, pod := range pods {
		podInfos = append(podInfos, shared.PodInfo{Name: pod.Name, Namespace: pod.Namespace})
	}
	tapStatus := shared.TapStatus{Pods: podInfos}

	if jsonValue, err := json.Marshal(tapStatus); err != nil {
		return fmt.Errorf("failed Marshal the tapped pods %w", err)
	} else {
		if response, err := http.Post(tappedPodsUrl, "application/json", bytes.NewBuffer(jsonValue)); err != nil {
			return fmt.Errorf("failed sending to API server the tapped pods %w", err)
		} else if response.StatusCode != 200 {
			return fmt.Errorf("failed sending to API server the tapped pods, response status code %v", response.StatusCode)
		} else {
			logger.Log.Debugf("Reported to server API about %d taped pods successfully", len(podInfos))
			return nil
		}
	}
}

func (provider *apiServerProvider) RequestSyncEntries(envName string, workspace string, sleepIntervalSec int, token string) error {
	if !provider.isReady {
		return fmt.Errorf("trying to reach api server when not initialized yet")
	}
	urlPath := fmt.Sprintf("%s/api/syncEntries?env=%s&workspace=%s&token=%s&interval=%v", provider.url, url.QueryEscape(envName), url.QueryEscape(workspace), url.QueryEscape(token), sleepIntervalSec)
	syncEntriesUrl, parseErr := url.ParseRequestURI(urlPath)
	if parseErr != nil {
		logger.Log.Fatal("Failed parsing the URL (consider changing the env name), err: %v", parseErr)
	}

	logger.Log.Debugf("Sync entries url %v", syncEntriesUrl.String())
	if response, requestErr := http.Get(syncEntriesUrl.String()); requestErr != nil {
		return fmt.Errorf("failed to notify api server for sync entries, err: %w", requestErr)
	} else if response.StatusCode != 200 {
		return fmt.Errorf("failed to notify api server for sync entries, status code: %v", response.StatusCode)
	} else {
		logger.Log.Infof(uiUtils.Purple, "Entries are syncing to UP9 for further analysis")
		return nil
	}
}

func (provider *apiServerProvider) GetGeneralStats() (map[string]interface{}, error) {
	if !provider.isReady {
		return nil, fmt.Errorf("trying to reach api server when not initialized yet")
	}
	generalStatsUrl := fmt.Sprintf("%s/api/generalStats", provider.url)

	response, requestErr := http.Get(generalStatsUrl)
	if requestErr != nil {
		return nil, fmt.Errorf("failed to get general stats for telemetry, err: %w", requestErr)
	} else if response.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get general stats for telemetry, status code: %v", response.StatusCode)
	}

	defer func() { _ = response.Body.Close() }()

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


func (provider *apiServerProvider) GetVersion() (string, error) {
	if !provider.isReady {
		return "", fmt.Errorf("trying to reach api server when not initialized yet")
	}
	versionUrl, _ := url.Parse(fmt.Sprintf("%s/metadata/version", provider.url))
	req := &http.Request{
		Method: http.MethodGet,
		URL:    versionUrl,
	}
	statusResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer statusResp.Body.Close()

	versionResponse := &shared.VersionResponse{}
	if err := json.NewDecoder(statusResp.Body).Decode(&versionResponse); err != nil {
		return "", err
	}

	return versionResponse.SemVer, nil
}
