package configStructs

const (
	RegenerateConfigName = "regenerate"
)

type ConfigConfig struct {
	Regenerate bool `yaml:"regenerate,omitempty" json:"regenerate,omitempty" default:"false" readonly:""`
}
