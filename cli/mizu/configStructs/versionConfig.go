package configStructs

const (
	DebugInfoVersionName = "debug"
)

type VersionConfig struct {
	DebugInfo bool `yaml:"debug" default:"false"`
}
