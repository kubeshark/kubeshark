package configStructs

const (
	DirectoryFetchName     = "directory"
	FromTimestampFetchName = "from"
	ToTimestampFetchName   = "to"
	GuiPortFetchName       = "gui-port"
)

type FetchConfig struct {
	Directory     string `yaml:"directory" default:"."`
	FromTimestamp int    `yaml:"from" default:"0"`
	ToTimestamp   int    `yaml:"to" default:"0"`
	GuiPort       uint16 `yaml:"gui-port" default:"8899"`
}
