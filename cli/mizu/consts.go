package mizu

import (
	"os"
	"path"
)

var (
	SemVer         = "0.0.1"
	Branch         = "develop"
	GitCommitHash  = "" // this var is overridden using ldflags in makefile when building
	BuildTimestamp = "" // this var is overridden using ldflags in makefile when building
	RBACVersion    = "v1"
)

const (
	ApiServerPodName       = "mizu-api-server"
	ClusterRoleBindingName = "mizu-cluster-role-binding"
	ClusterRoleName        = "mizu-cluster-role"
	K8sAllNamespaces       = ""
	RoleBindingName        = "mizu-role-binding"
	RoleName               = "mizu-role"
	ServiceAccountName     = "mizu-service-account"
	TapperDaemonSetName    = "mizu-tapper-daemon-set"
	TapperPodName          = "mizu-tapper"
	ConfigMapName          = "mizu-policy"
)

func GetMizuFolderPath() string {
	home, homeDirErr := os.UserHomeDir()
	if homeDirErr != nil {
		return ""
	}
	return path.Join(home, ".mizu")
}
