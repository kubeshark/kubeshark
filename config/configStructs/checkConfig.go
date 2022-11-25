package configStructs

const (
	PreTapCheckName     = "pre-tap"
	PreInstallCheckName = "pre-install"
	ImagePullCheckName  = "image-pull"
)

type CheckConfig struct {
	PreTap     bool `yaml:"pre-tap"`
	PreInstall bool `yaml:"pre-install"`
	ImagePull  bool `yaml:"image-pull"`
}
