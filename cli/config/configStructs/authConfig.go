package configStructs

type AuthConfig struct {
	EnvName      string    `yaml:"env-name" default:"up9.app"`
	Token        string    `yaml:"token"`
}
