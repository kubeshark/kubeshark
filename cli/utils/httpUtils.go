package utils

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

// Get - When err is nil, resp always contains a non-nil resp.Body.
// Caller should close resp.Body when done reading from it.
func Get(url string, client *http.Client) (*http.Response, error) {
	return checkError(client.Get(url))
}

// Post - When err is nil, resp always contains a non-nil resp.Body.
// Caller should close resp.Body when done reading from it.
func Post(url, contentType string, body io.Reader, client *http.Client) (*http.Response, error) {
	return checkError(client.Post(url, contentType, body))
}

// Do - When err is nil, resp always contains a non-nil resp.Body.
// Caller should close resp.Body when done reading from it.
func Do(req *http.Request, client *http.Client) (*http.Response, error) {
	return checkError(client.Do(req))
}

func checkError(response *http.Response, errInOperation error) (*http.Response, error) {
	if errInOperation != nil {
		return response, errInOperation
		// Check only if status != 200 (and not status >= 300). Agent APIs return only 200 on success.
	} else if response.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(response.Body)
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
