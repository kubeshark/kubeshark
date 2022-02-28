package configStructs

const (
	PreTapCheckName = "pre-tap"
)

type CheckConfig struct {
	PreTap bool `yaml:"pre-tap"`
}
