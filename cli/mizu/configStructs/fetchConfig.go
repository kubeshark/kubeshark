package configStructs

const (
	DirectoryFetchName     = "directory"
	FromTimestampFetchName = "from"
	ToTimestampFetchName   = "to"
	MizuPortFetchName      = "port"
)

type FetchConfig struct {
	Directory     string `yaml:"directory" default:"."`
	FromTimestamp int    `yaml:"from" default:"0"`
	ToTimestamp   int    `yaml:"to" default:"0"`
	MizuPort      uint16 `yaml:"port" default:"8899"`
}
