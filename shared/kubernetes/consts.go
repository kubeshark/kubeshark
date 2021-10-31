package kubernetes

const (
	MizuResourcesPrefix        = "mizu-"
	ApiServerPodName           = MizuResourcesPrefix + "api-server"
	ClusterRoleBindingName     = MizuResourcesPrefix + "cluster-role-binding"
	ClusterRoleName            = MizuResourcesPrefix + "cluster-role"
	K8sAllNamespaces           = ""
	RoleBindingName            = MizuResourcesPrefix + "role-binding"
	RoleName                   = MizuResourcesPrefix + "role"
	ServiceAccountName         = MizuResourcesPrefix + "service-account"
	TapperDaemonSetName        = MizuResourcesPrefix + "tapper-daemon-set"
	TapperPodName              = MizuResourcesPrefix + "tapper"
	ConfigMapName              = MizuResourcesPrefix + "config"
	MinKubernetesServerVersion = "1.16.0"
)

