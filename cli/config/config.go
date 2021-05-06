package config

type Options struct {
	Quiet          bool
	NoGUI          bool
	GuiPort        uint16
	Namespace      string
	AllNamespaces  bool
	KubeConfigPath string
	MizuImage      string
	MizuPodPort    uint16
}

var Configuration = &Options{}
