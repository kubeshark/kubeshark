package kubernetes

import (
	"context"
	"fmt"
	"sync"
	"time"

	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// HubTokenAudience must match the hub's internalauth TokenAudience.
	HubTokenAudience = "kubeshark-hub"
	// CLIServiceAccountName is the ServiceAccount the CLI mints a token for;
	// must match the SA the helm chart creates and the hub's
	// AUTH_CLI_SERVICE_ACCOUNTS allowlist.
	CLIServiceAccountName = "kubeshark-cli"

	hubTokenExpirySeconds = 3600
	// hubTokenRenewMargin re-mints this long before expiry so requests in
	// flight always carry a comfortably valid token.
	hubTokenRenewMargin = 5 * time.Minute
)

// MintHubToken requests a short-lived ServiceAccount token (audience
// HubTokenAudience) for the CLI ServiceAccount via the TokenRequest API. The
// hub validates it with TokenReview and maps the SA to a role. The caller's
// kube RBAC must permit `create` on serviceaccounts/token for that SA — which
// is what gates who may use the CLI against a gated Hub.
func (provider *Provider) MintHubToken(ctx context.Context, namespace string) (string, error) {
	tok, _, err := provider.mintHubToken(ctx, namespace)
	return tok, err
}

// mintHubToken is MintHubToken plus the token's expiry (from the TokenRequest
// response, so a cluster that caps the duration below the request is respected;
// falls back to the requested TTL if the API omits it).
func (provider *Provider) mintHubToken(ctx context.Context, namespace string) (string, time.Time, error) {
	exp := int64(hubTokenExpirySeconds)
	tr, err := provider.clientSet.CoreV1().ServiceAccounts(namespace).CreateToken(
		ctx,
		CLIServiceAccountName,
		&authenticationv1.TokenRequest{
			Spec: authenticationv1.TokenRequestSpec{
				Audiences:         []string{HubTokenAudience},
				ExpirationSeconds: &exp,
			},
		},
		metav1.CreateOptions{},
	)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("minting hub token for serviceaccount %s/%s: %w", namespace, CLIServiceAccountName, err)
	}
	expireAt := tr.Status.ExpirationTimestamp.Time
	if expireAt.IsZero() {
		expireAt = time.Now().Add(time.Duration(hubTokenExpirySeconds) * time.Second)
	}
	return tr.Status.Token, expireAt, nil
}

// HubTokenRenewer hands out a hub SA token, transparently re-minting it before
// it expires. Safe for concurrent use; intended for long-lived CLI processes
// (mcp/console proxy mode) that would otherwise 401 after the ~1h token TTL.
type HubTokenRenewer struct {
	provider  *Provider
	namespace string

	mu       sync.Mutex
	token    string
	expireAt time.Time
}

// NewHubTokenRenewer builds a renewer that mints kubeshark-cli tokens in the
// given namespace.
func NewHubTokenRenewer(provider *Provider, namespace string) *HubTokenRenewer {
	return &HubTokenRenewer{provider: provider, namespace: namespace}
}

// Token returns a currently-valid hub token, minting or re-minting as needed.
// On a mint failure it returns the last token (possibly "") so the caller can
// fall back to the License-Key rather than hard-failing.
func (r *HubTokenRenewer) Token() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.token == "" || time.Now().After(r.expireAt.Add(-hubTokenRenewMargin)) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		tok, expireAt, err := r.provider.mintHubToken(ctx, r.namespace)
		cancel()
		if err != nil {
			return r.token
		}
		r.token, r.expireAt = tok, expireAt
	}
	return r.token
}
