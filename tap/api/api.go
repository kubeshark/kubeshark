package api

import "plugin"

type Extension struct {
	Name      string
	Path      string
	Plug      *plugin.Plugin
	Port      string
	Dissector Dissector
}

type Dissector interface {
	Register(*Extension)
	Ping()
}
