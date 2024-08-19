package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	X_KUBESHARK_CAPTURE_HEADER_KEY          = "X-Kubeshark-Capture"
	X_KUBESHARK_CAPTURE_HEADER_IGNORE_VALUE = "ignore"
)

// Get - When err is nil, resp always contains a non-nil resp.Body.
// Caller should close resp.Body when done reading from it.
func Get(url string, client *http.Client) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	AddIgnoreCaptureHeader(req)

	return checkError(client.Do(req))
}

// Post - When err is nil, resp always contains a non-nil resp.Body.
// Caller should close resp.Body when done reading from it.
func Post(url, contentType string, body io.Reader, client *http.Client, licenseKey string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	AddIgnoreCaptureHeader(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("License-Key", licenseKey)

	return checkError(client.Do(req))
}

// Do - When err is nil, resp always contains a non-nil resp.Body.
// Caller should close resp.Body when done reading from it.
func Do(req *http.Request, client *http.Client) (*http.Response, error) {
	return checkError(client.Do(req))
}

func checkError(response *http.Response, errInOperation error) (*http.Response, error) {
	if errInOperation != nil {
		return response, errInOperation
		// Check only if status != 200 (and not status >= 300). Hub return only 200 on success.
	} else if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(response.Body)
		response.Body.Close()
		response.Body = io.NopCloser(bytes.NewBuffer(body)) // rewind
		if err != nil {
			return response, err
		}

		errorMsg := strings.ReplaceAll(string(body), "\n", ";")
		return response, fmt.Errorf("got response with status code: %d, body: %s", response.StatusCode, errorMsg)
	}

	return response, nil
}

func AddIgnoreCaptureHeader(req *http.Request) {
	req.Header.Set(X_KUBESHARK_CAPTURE_HEADER_KEY, X_KUBESHARK_CAPTURE_HEADER_IGNORE_VALUE)
}
