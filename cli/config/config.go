package config

type Options struct {
	Quiet          bool
	NoDashboard    bool
	DashboardPort  uint16
	Namespace      string
	AllNamespaces  bool
	KubeConfigPath string
	MizuImage      string
	MizuPodPort    uint16
	TappedPodName  string
}

var Configuration = &Options{}
