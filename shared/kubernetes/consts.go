package kubernetes

const (
	MizuResourcesPrefix          = "mizu-"
	ApiServerPodName             = MizuResourcesPrefix + "api-server"
	ClusterRoleBindingName       = MizuResourcesPrefix + "cluster-role-binding"
	DaemonClusterRoleBindingName = MizuResourcesPrefix + "cluster-role-binding-daemon"
	ClusterRoleName              = MizuResourcesPrefix + "cluster-role"
	DaemonClusterRoleName        = MizuResourcesPrefix + "cluster-role-daemon"
	K8sAllNamespaces             = ""
	RoleBindingName              = MizuResourcesPrefix + "role-binding"
	RoleName                     = MizuResourcesPrefix + "role"
	ServiceAccountName           = MizuResourcesPrefix + "service-account"
	TapperDaemonSetName          = MizuResourcesPrefix + "tapper-daemon-set"
	TapperPodName                = MizuResourcesPrefix + "tapper"
	ConfigMapName                = MizuResourcesPrefix + "config"
	PersistentVolumeClaimName    = MizuResourcesPrefix + "volume-claim"
	MinKubernetesServerVersion   = "1.16.0"
)
