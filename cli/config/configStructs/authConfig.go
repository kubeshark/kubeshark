package configStructs

import (
	"errors"
	"fmt"
	"time"
)

const (
	ClientIdAuthName     = "client-id"
	ClientSecretAuthName = "client-secret"
)

type AuthConfig struct {
	EnvName      string    `yaml:"env-name" default:"up9.app"`
	ClientId     string    `yaml:"client-id"`
	ClientSecret string    `yaml:"client-secret"`
	Token        string    `yaml:"token"`
	ExpiryDate   time.Time `yaml:"expiry-date"`
}

func (config *AuthConfig) Validate() error {
	if config.ClientId != "" && config.ClientSecret == "" {
		return errors.New(fmt.Sprintf("--%s must also be provided when using --%s", ClientSecretAuthName, ClientIdAuthName))
	}

	if config.ClientSecret != "" && config.ClientId == "" {
		return errors.New(fmt.Sprintf("--%s must also be provided when using --%s", ClientIdAuthName, ClientSecretAuthName))
	}

	return nil
}
