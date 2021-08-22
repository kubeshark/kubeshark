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
	MizuResourcesPrefix    = "mizu-"
	ApiServerPodName       = MizuResourcesPrefix + "api-server"
	ClusterRoleBindingName = MizuResourcesPrefix + "cluster-role-binding"
	ClusterRoleName        = MizuResourcesPrefix + "cluster-role"
	K8sAllNamespaces       = ""
	RoleBindingName        = MizuResourcesPrefix + "role-binding"
	RoleName               = MizuResourcesPrefix + "role"
	ServiceAccountName     = MizuResourcesPrefix + "service-account"
	TapperDaemonSetName    = MizuResourcesPrefix + "tapper-daemon-set"
	TapperPodName          = MizuResourcesPrefix + "tapper"
	ConfigMapName          = MizuResourcesPrefix + "policy"
)

func GetMizuFolderPath() string {
	home, homeDirErr := os.UserHomeDir()
	if homeDirErr != nil {
		return ""
	}
	return path.Join(home, ".mizu")
}
