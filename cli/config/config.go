package config

type Options struct {
	DisplayVersion   bool
	Quiet            bool
	NoDashboard      bool
	DashboardPort    uint16
	Namespace        string
	AllNamespaces    bool
	KubeConfigPath   string
}

var Configuration = &Options{}
