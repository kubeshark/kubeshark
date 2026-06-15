package kubernetes

import (
	"context"
	"fmt"

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
)

// MintHubToken requests a short-lived ServiceAccount token (audience
// HubTokenAudience) for the CLI ServiceAccount via the TokenRequest API. The
// hub validates it with TokenReview and maps the SA to a role. The caller's
// kube RBAC must permit `create` on serviceaccounts/token for that SA — which
// is what gates who may use the CLI against a gated Hub.
func (provider *Provider) MintHubToken(ctx context.Context, namespace string) (string, error) {
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
		return "", fmt.Errorf("minting hub token for serviceaccount %s/%s: %w", namespace, CLIServiceAccountName, err)
	}
	return tr.Status.Token, nil
}
