package mizu

var (
	Version       = "v0.0.1"
	Branch        = "develop"
	GitCommitHash = "" // this var is overridden using ldflags in makefile when building
)

const (
	ResourcesNamespace  = "default"
	TapperDaemonSetName = "mizu-tapper-daemon-set"
	AggregatorPodName   = "mizu-collector"
	TapperPodName       = "mizu-tapper"
	K8sAllNamespaces    = ""
)
