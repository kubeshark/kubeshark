package api

import "plugin"

type Extension struct {
	Name string
	Path string
	Plug *plugin.Plugin
	Port int
}

type Dissector interface {
	Register(*Extension)
}
