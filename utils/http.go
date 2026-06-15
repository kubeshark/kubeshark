package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	X_KUBESHARK_CAPTURE_HEADER_KEY          = "X-Kubeshark-Capture"
	X_KUBESHARK_CAPTURE_HEADER_IGNORE_VALUE = "ignore"
	LICENSE_KEY_HEADER                      = "License-Key"
	// CLI_AUTH_HEADER carries a ServiceAccount bearer token to the Hub.
	// A custom (non-Authorization) header so it survives the kube
	// API-server service proxy, which consumes Authorization.
	CLI_AUTH_HEADER = "X-Kubeshark-Authorization"
)

// ErrHubAuthRequired indicates the Hub rejected the request because auth is
// enabled but no valid credential was presented (missing/expired license).
var ErrHubAuthRequired = errors.New("hub requires authentication: set a valid license (config 'license') or credentials")

// hubAuthRoundTripper attaches the CLI's Hub credential to every request so any
// client built with it authenticates to a gated Hub. It prefers a scoped
// ServiceAccount token (CLI_AUTH_HEADER) when present, falling back to the
// License-Key (admin/transitional). Both are custom headers so they survive
// the kube API-server service proxy, unlike an Authorization bearer.
type hubAuthRoundTripper struct {
	saToken    string
	licenseKey string
	base       http.RoundTripper
}

func (rt *hubAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	base := rt.base
	if base == nil {
		base = http.DefaultTransport
	}
	switch {
	case rt.saToken != "":
		if req.Header.Get(CLI_AUTH_HEADER) == "" {
			req = req.Clone(req.Context())
			req.Header.Set(CLI_AUTH_HEADER, rt.saToken)
		}
	case rt.licenseKey != "":
		if req.Header.Get(LICENSE_KEY_HEADER) == "" {
			req = req.Clone(req.Context())
			req.Header.Set(LICENSE_KEY_HEADER, rt.licenseKey)
		}
	}
	return base.RoundTrip(req)
}

// NewHubHTTPClient returns an *http.Client that authenticates to the Hub with
// the License-Key header (Phase 1).
func NewHubHTTPClient(timeout time.Duration, licenseKey string) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: &hubAuthRoundTripper{licenseKey: licenseKey},
	}
}

// NewHubHTTPClientWithToken returns an *http.Client that authenticates to the
// Hub with a ServiceAccount token when saToken is set, otherwise the
// License-Key.
func NewHubHTTPClientWithToken(timeout time.Duration, saToken, licenseKey string) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: &hubAuthRoundTripper{saToken: saToken, licenseKey: licenseKey},
	}
}

// HubAuthTransport wraps base so requests carry the License-Key header. Use
// when a client needs custom transport settings (e.g. streaming downloads)
// but must still authenticate to the Hub.
func HubAuthTransport(licenseKey string, base http.RoundTripper) http.RoundTripper {
	return &hubAuthRoundTripper{licenseKey: licenseKey, base: base}
}

// HubAuthTransportWithToken is HubAuthTransport with a ServiceAccount token
// (preferred over the License-Key when set).
func HubAuthTransportWithToken(saToken, licenseKey string, base http.RoundTripper) http.RoundTripper {
	return &hubAuthRoundTripper{saToken: saToken, licenseKey: licenseKey, base: base}
}

// IsAuthRequired reports whether the response indicates the Hub demanded
// authentication — a 401, or a 302 redirect to an SSO login page.
func IsAuthRequired(resp *http.Response) bool {
	return resp != nil && (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusFound)
}

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
