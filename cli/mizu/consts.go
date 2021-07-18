package mizu

var (
	SemVer         = "0.0.1"
	Branch         = "develop"
	GitCommitHash  = "" // this var is overridden using ldflags in makefile when building
	BuildTimestamp = "" // this var is overridden using ldflags in makefile when building
	RBACVersion    = "v1"
)

const (
	AggregatorPodName       = "mizu-collector"
	ClusterRoleBindingsName = "mizu-cluster-role-binding"
	ClusterRoleName         = "mizu-cluster-role"
	K8sAllNamespaces        = ""
	ResourcesNamespace      = "mizu"
	ServiceAccountName      = "mizu-service-account"
	TapperDaemonSetName     = "mizu-tapper-daemon-set"
	TapperPodName           = "mizu-tapper"
)

const (
	Black        = "\033[1;30m%s\033[0m"
	Red          = "\033[1;31m%s\033[0m"
	Green        = "\033[1;32m%s\033[0m"
	Yellow       = "\033[1;33m%s\033[0m"
	Purple       = "\033[1;34m%s\033[0m"
	Magenta      = "\033[1;35m%s\033[0m"
	Teal         = "\033[1;36m%s\033[0m"
	White        = "\033[1;37m%s\033[0m"
)
