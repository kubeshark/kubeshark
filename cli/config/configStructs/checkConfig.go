package configStructs

const (
	PreTapCheckName    = "pre-tap"
	ImagePullCheckName = "image-pull"
)

type CheckConfig struct {
	PreTap    bool `yaml:"pre-tap"`
	ImagePull bool `yaml:"image-pull"`
}
