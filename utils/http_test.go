package utils

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestHubHTTPClientRedirects(t *testing.T) {
	factories := []struct {
		name      string
		token     string
		newClient func() *http.Client
	}{
		{"license", "", func() *http.Client { return NewHubHTTPClient(time.Second, "test-license") }},
		{"fixed token", "test-token", func() *http.Client {
			return NewHubHTTPClientWithToken(time.Second, "test-token", "test-license")
		}},
		{"fixed token fallback", "", func() *http.Client {
			return NewHubHTTPClientWithToken(time.Second, "", "test-license")
		}},
		{"token source", "test-token", func() *http.Client {
			return NewHubHTTPClientWithTokenSource(time.Second, func() string { return "test-token" }, "test-license")
		}},
		{"token source fallback", "", func() *http.Client {
			return NewHubHTTPClientWithTokenSource(time.Second, func() string { return "" }, "test-license")
		}},
		{"download token", "test-token", func() *http.Client {
			return &http.Client{
				CheckRedirect: StopOnSSORedirect,
				Transport:     HubAuthTransportWithTokenSource(func() string { return "test-token" }, "test-license", http.DefaultTransport),
			}
		}},
		{"download fallback", "", func() *http.Client {
			return &http.Client{
				CheckRedirect: StopOnSSORedirect,
				Transport:     HubAuthTransportWithTokenSource(nil, "test-license", http.DefaultTransport),
			}
		}},
	}
	for _, factory := range factories {
		t.Run(factory.name, func(t *testing.T) {
			for _, status := range []int{http.StatusMovedPermanently, http.StatusTemporaryRedirect, http.StatusPermanentRedirect} {
				t.Run(http.StatusText(status), func(t *testing.T) {
					for _, crossOrigin := range []bool{false, true} {
						name := "same origin"
						if crossOrigin {
							name = "different port"
						}
						t.Run(name, func(t *testing.T) {
							var destinationRequests atomic.Int32
							destination := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
								destinationRequests.Add(1)
								if got := r.Header.Get(CLI_AUTH_HEADER); got != factory.token {
									t.Errorf("token = %q, want %q", got, factory.token)
								}
								wantLicense := ""
								if factory.token == "" {
									wantLicense = "test-license"
								}
								if got := r.Header.Get(LICENSE_KEY_HEADER); got != wantLicense {
									t.Errorf("license = %q, want %q", got, wantLicense)
								}
								if status == http.StatusTemporaryRedirect || status == http.StatusPermanentRedirect {
									body, err := io.ReadAll(r.Body)
									if err != nil || r.Method != http.MethodPost || string(body) != "tool-call" {
										t.Errorf("redirected tool call: method=%s body=%q err=%v", r.Method, body, err)
									}
								}
								w.WriteHeader(http.StatusNoContent)
							})
							target := httptest.NewServer(destination)
							defer target.Close()
							hub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
								if r.URL.Path == "/done" {
									destination.ServeHTTP(w, r)
									return
								}
								location := "/done"
								if crossOrigin {
									location = target.URL + "/done"
								}
								http.Redirect(w, r, location, status)
							}))
							defer hub.Close()

							resp, err := factory.newClient().Post(hub.URL+"/redirect", "application/json", strings.NewReader("tool-call"))
							if resp != nil {
								defer resp.Body.Close()
							}
							if crossOrigin {
								if err == nil || !strings.Contains(err.Error(), "outside the Hub origin") {
									t.Fatalf("cross-origin request error = %v", err)
								}
								if got := destinationRequests.Load(); got != 0 {
									t.Fatalf("cross-origin destination received %d requests", got)
								}
								return
							}
							if err != nil {
								t.Fatal(err)
							}
							if resp.StatusCode != http.StatusNoContent || destinationRequests.Load() != 1 {
								t.Fatalf("same-origin redirect: status=%d destination requests=%d", resp.StatusCode, destinationRequests.Load())
							}
						})
					}
				})
			}
		})
	}
}

func TestStopOnSSORedirectOrigin(t *testing.T) {
	tests := []struct {
		name    string
		from    string
		to      string
		status  int
		allow   bool
		wantSSO bool
	}{
		{"different hostname", "https://hub.example/api", "https://other.example/api", 307, false, false},
		{"subdomain", "https://hub.example/api", "https://child.hub.example/api", 307, false, false},
		{"different port", "https://hub.example/api", "https://hub.example:444/api", 308, false, false},
		{"HTTPS downgrade", "https://hub.example/api", "http://hub.example:443/api", 301, false, false},
		{"HTTP upgrade", "http://hub.example/api", "https://hub.example:80/api", 301, false, false},
		{"default HTTPS port", "https://hub.example/api", "https://hub.example:443/done", 307, true, false},
		{"default HTTP port", "http://hub.example:80/api", "http://hub.example/done", 308, true, false},
		{"hostname case", "https://HUB.example/api", "https://hub.EXAMPLE/done", 301, true, false},
		{"IPv6 default port", "http://[::1]:80/api", "http://[::1]/done", 307, true, false},
		{"same-origin SSO", "https://hub.example/api", "https://hub.example/login", 302, false, true},
		{"external SSO", "https://hub.example/api", "https://login.example/", 303, false, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			original, err := http.NewRequest(http.MethodGet, test.from, nil)
			if err != nil {
				t.Fatal(err)
			}
			redirect, err := http.NewRequest(http.MethodGet, test.to, nil)
			if err != nil {
				t.Fatal(err)
			}
			redirect.Response = &http.Response{StatusCode: test.status}
			err = StopOnSSORedirect(redirect, []*http.Request{original})
			if test.allow {
				if err != nil {
					t.Fatalf("same-origin redirect rejected: %v", err)
				}
			} else if test.wantSSO {
				if err != http.ErrUseLastResponse || !IsAuthRequired(redirect.Response) {
					t.Fatalf("SSO response not preserved: %v", err)
				}
			} else if err == nil || err == http.ErrUseLastResponse {
				t.Fatalf("off-origin redirect not rejected with an error: %v", err)
			}
		})
	}
}
