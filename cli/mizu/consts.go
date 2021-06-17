package mizu

var (
	CalVer         = "yy.MM.DD.HH.mm.ss" // this var is overridden using ldflags in makefile when building
	Branch         = "develop"
	GitCommitHash  = "" // this var is overridden using ldflags in makefile when building
	BuildTimeUTC   = "" // this var is overridden using ldflags in makefile when building
	BuildTimestamp = "" // this var is overridden using ldflags in makefile when building
	RBACVersion    = "v1"
)

const (
	ResourcesNamespace  = "default"
	TapperDaemonSetName = "mizu-tapper-daemon-set"
	AggregatorPodName   = "mizu-collector"
	TapperPodName       = "mizu-tapper"
	K8sAllNamespaces    = ""
)
