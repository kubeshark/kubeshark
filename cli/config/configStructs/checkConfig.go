package configStructs

const (
	PreTapCheckName             = "pre-tap"
	ImagesConnectivityCheckName = "images-connectivity"
)

type CheckConfig struct {
	PreTap             bool `yaml:"pre-tap"`
	ImagesConnectivity bool `yaml:"images-connectivity"`
}
