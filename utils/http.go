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
//
// The SA token is sourced via saTokenFunc on every request rather than captured
// once, so a caller with cluster access (proxy mode) can hand in a renewing
// source and long-lived clients keep working past the token's ~1h expiry.
type hubAuthRoundTripper struct {
	saTokenFunc func() string
	licenseKey  string
	base        http.RoundTripper
}

func (rt *hubAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	base := rt.base
	if base == nil {
		base = http.DefaultTransport
	}
	saToken := ""
	if rt.saTokenFunc != nil {
		saToken = rt.saTokenFunc()
	}
	switch {
	case saToken != "":
		if req.Header.Get(CLI_AUTH_HEADER) == "" {
			req = req.Clone(req.Context())
			req.Header.Set(CLI_AUTH_HEADER, saToken)
		}
	case rt.licenseKey != "":
		if req.Header.Get(LICENSE_KEY_HEADER) == "" {
			req = req.Clone(req.Context())
			req.Header.Set(LICENSE_KEY_HEADER, rt.licenseKey)
		}
	}
	return base.RoundTrip(req)
}

// staticToken adapts a fixed token string to the saTokenFunc source, returning
// nil for an empty token so the round-tripper falls back to the License-Key.
func staticToken(saToken string) func() string {
	if saToken == "" {
		return nil
	}
	return func() string { return saToken }
}

// StopOnSSORedirect is an *http.Client CheckRedirect that does NOT follow
// SSO-style auth redirects (302 Found / 303 See Other) — it returns the
// redirect response so callers can detect auth-required via IsAuthRequired
// instead of silently following it to an HTML login page (and, for downloads,
// writing that page to disk). Other redirects (301/307/308) are still followed.
func StopOnSSORedirect(req *http.Request, _ []*http.Request) error {
	if req.Response != nil {
		switch req.Response.StatusCode {
		case http.StatusFound, http.StatusSeeOther:
			return http.ErrUseLastResponse
		}
	}
	return nil
}

// NewHubHTTPClient returns an *http.Client that authenticates to the Hub with
// the License-Key header.
func NewHubHTTPClient(timeout time.Duration, licenseKey string) *http.Client {
	return &http.Client{
		Timeout:       timeout,
		CheckRedirect: StopOnSSORedirect,
		Transport:     &hubAuthRoundTripper{licenseKey: licenseKey},
	}
}

// NewHubHTTPClientWithToken returns an *http.Client that authenticates to the
// Hub with a fixed ServiceAccount token when saToken is set, otherwise the
// License-Key.
func NewHubHTTPClientWithToken(timeout time.Duration, saToken, licenseKey string) *http.Client {
	return &http.Client{
		Timeout:       timeout,
		CheckRedirect: StopOnSSORedirect,
		Transport:     &hubAuthRoundTripper{saTokenFunc: staticToken(saToken), licenseKey: licenseKey},
	}
}

// NewHubHTTPClientWithTokenSource is NewHubHTTPClientWithToken with a token
// source consulted per request, so a renewing source keeps the client
// authenticated past the token's expiry. A nil source falls back to the
// License-Key.
func NewHubHTTPClientWithTokenSource(timeout time.Duration, saTokenFunc func() string, licenseKey string) *http.Client {
	return &http.Client{
		Timeout:       timeout,
		CheckRedirect: StopOnSSORedirect,
		Transport:     &hubAuthRoundTripper{saTokenFunc: saTokenFunc, licenseKey: licenseKey},
	}
}

// HubAuthTransport wraps base so requests carry the License-Key header. Use
// when a client needs custom transport settings (e.g. streaming downloads)
// but must still authenticate to the Hub.
func HubAuthTransport(licenseKey string, base http.RoundTripper) http.RoundTripper {
	return &hubAuthRoundTripper{licenseKey: licenseKey, base: base}
}

// HubAuthTransportWithToken is HubAuthTransport with a fixed ServiceAccount
// token (preferred over the License-Key when set).
func HubAuthTransportWithToken(saToken, licenseKey string, base http.RoundTripper) http.RoundTripper {
	return &hubAuthRoundTripper{saTokenFunc: staticToken(saToken), licenseKey: licenseKey, base: base}
}

// HubAuthTransportWithTokenSource is HubAuthTransportWithToken with a token
// source consulted per request (see NewHubHTTPClientWithTokenSource).
func HubAuthTransportWithTokenSource(saTokenFunc func() string, licenseKey string, base http.RoundTripper) http.RoundTripper {
	return &hubAuthRoundTripper{saTokenFunc: saTokenFunc, licenseKey: licenseKey, base: base}
}

// IsAuthRequired reports whether the response indicates the Hub demanded
// authentication — a 401, or a 302/303 redirect to an SSO login page (the
// hub clients are configured not to follow those; see StopOnSSORedirect).
func IsAuthRequired(resp *http.Response) bool {
	return resp != nil && (resp.StatusCode == http.StatusUnauthorized ||
		resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusSeeOther)
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
