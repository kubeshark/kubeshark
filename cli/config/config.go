package config

type Options struct {
	NoGUI          bool
	GuiPort        uint16
	Namespace      string
	KubeConfigPath string
	MizuImage      string
	MizuPodPort    uint16
}

var Configuration = &Options{}
