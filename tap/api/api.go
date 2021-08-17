package api

import (
	"bufio"
	"plugin"
)

type Extension struct {
	Name      string
	Path      string
	Plug      *plugin.Plugin
	Ports     []string
	Dissector Dissector
}

type Dissector interface {
	Register(*Extension)
	Ping()
	Dissect(b *bufio.Reader) interface{}
}
