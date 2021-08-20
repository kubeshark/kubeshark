package apiserver

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
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
}

var Provider = apiServerProvider{}

func (provider *apiServerProvider) Init(url string, retries int) error {
	healthUrl := fmt.Sprintf("%s/", url)
	retriesLeft := retries
	for retriesLeft > 0 {
		if response, err := http.Get(healthUrl); err != nil {
			logger.Log.Debugf("[ERROR] failed connecting to api server %v", err)
		} else if response.StatusCode != 200 {
			logger.Log.Debugf("can't connect to api server yet, response status code %v", response.StatusCode)
		} else {
			logger.Log.Debugf("connection test to api server passed successfully")
			break
		}
		retriesLeft -= 1
		time.Sleep(time.Second)
	}

	if retriesLeft == 0 {
		return fmt.Errorf("couldn't reach the api server after %v retries", retries)
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

func (provider *apiServerProvider) RequestAnalysis(analysisDestination string, sleepIntervalSec int) error {
	if !provider.isReady {
		return fmt.Errorf("trying to reach api server when not initialized yet")
	}
	urlPath := fmt.Sprintf("http://%s/api/uploadEntries?dest=%s&interval=%v", provider.url, url.QueryEscape(analysisDestination), sleepIntervalSec)
	u, parseErr := url.ParseRequestURI(urlPath)
	if parseErr != nil {
		logger.Log.Fatal("Failed parsing the URL (consider changing the analysis dest URL), err: %v", parseErr)
	}

	logger.Log.Debugf("Analysis url %v", u.String())
	if response, requestErr := http.Get(u.String()); requestErr != nil {
		return fmt.Errorf("failed to notify agent for analysis, err: %w", requestErr)
	} else if response.StatusCode != 200 {
		return fmt.Errorf("failed to notify agent for analysis, status code: %v", response.StatusCode)
	} else {
		logger.Log.Infof(uiUtils.Purple, "Traffic is uploading to UP9 for further analysis")
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

func (provider *apiServerProvider) GetHars(fromTimestamp int, toTimestamp int) (*zip.Reader, error) {
	if !provider.isReady {
		return nil, fmt.Errorf("trying to reach api server when not initialized yet")
	}
	resp, err := http.Get(fmt.Sprintf("%s/api/har?from=%v&to=%v", provider.url, fromTimestamp, toTimestamp))
	if err != nil {
		return nil, fmt.Errorf("failed getting har from api server %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading hars %w", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, fmt.Errorf("failed craeting zip reader %w", err)
	}
	return zipReader, nil
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
