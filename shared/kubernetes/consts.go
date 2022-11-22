package kubernetes

const (
	KubesharkResourcesPrefix   = "kubeshark-"
	ApiServerPodName           = KubesharkResourcesPrefix + "api-server"
	ClusterRoleBindingName     = KubesharkResourcesPrefix + "cluster-role-binding"
	ClusterRoleName            = KubesharkResourcesPrefix + "cluster-role"
	K8sAllNamespaces           = ""
	RoleBindingName            = KubesharkResourcesPrefix + "role-binding"
	RoleName                   = KubesharkResourcesPrefix + "role"
	ServiceAccountName         = KubesharkResourcesPrefix + "service-account"
	TapperDaemonSetName        = KubesharkResourcesPrefix + "tapper-daemon-set"
	TapperPodName              = KubesharkResourcesPrefix + "tapper"
	ConfigMapName              = KubesharkResourcesPrefix + "config"
	MinKubernetesServerVersion = "1.16.0"
)

const (
	LabelPrefixApp           = "app.kubernetes.io/"
	LabelManagedBy           = LabelPrefixApp + "managed-by"
	LabelCreatedBy           = LabelPrefixApp + "created-by"
	LabelValueKubeshark      = "kubeshark"
	LabelValueKubesharkCLI   = "kubeshark-cli"
	LabelValueKubesharkAgent = "kubeshark-agent"
)
