package configStructs

const (
	ServerUrlCheckName = "server-url"
)

type CheckConfig struct {
	ServerUrl string `yaml:"server-url"`
}
