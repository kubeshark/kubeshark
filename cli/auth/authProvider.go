package auth

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

func LoginNonInteractively(clientId string, clientSecret string, envName string) (*oauth2.Token, error) {
	authConfig := clientcredentials.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		TokenURL:     fmt.Sprintf("https://auth.%s/auth/realms/testr/protocol/openid-connect/token", envName),
	}

	token, err := authConfig.Token(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed creating token, err: %v", err)
	}

	return token, nil
}
