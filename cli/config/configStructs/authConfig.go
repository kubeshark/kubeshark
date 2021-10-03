package configStructs

import (
	"time"
)

type AuthConfig struct {
	EnvName      string    `yaml:"env-name" default:"up9.app"`
	Token        string    `yaml:"token"`
	ExpiryDate   time.Time `yaml:"expiry-date"`
}
