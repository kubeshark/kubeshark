package configStructs

const (
	PreInstallCheckName = "pre-install"
	PreTapCheckName     = "pre-tap"
)

type CheckConfig struct {
	PreInstall bool `yaml:"pre-install"`
	PreTap     bool `yaml:"pre-tap"`
}
