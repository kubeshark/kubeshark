package mizu

var (
	Version = "v0.0.1"
	GitCommitHash = "" // this var is overridden using ldflags in makefile when building
)

const (
	MizuResourcesNamespace = "default"
	TapperDaemonSetName = "mizu-tapper-daemon-set"
	aggregatorPodName = "mizu-collector"
	tapperPodName = "mizu-tapper"
)
