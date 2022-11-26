package kubernetes

const (
	KubesharkResourcesPrefix   = "ks-"
	FrontPodName               = KubesharkResourcesPrefix + "front"
	FrontServiceName           = FrontPodName
	ApiServerPodName           = KubesharkResourcesPrefix + "hub"
	ApiServerServiceName       = ApiServerPodName
	ClusterRoleBindingName     = KubesharkResourcesPrefix + "cluster-role-binding"
	ClusterRoleName            = KubesharkResourcesPrefix + "cluster-role"
	K8sAllNamespaces           = ""
	RoleBindingName            = KubesharkResourcesPrefix + "role-binding"
	RoleName                   = KubesharkResourcesPrefix + "role"
	ServiceAccountName         = KubesharkResourcesPrefix + "service-account"
	TapperDaemonSetName        = KubesharkResourcesPrefix + "worker-daemon-set"
	TapperPodName              = KubesharkResourcesPrefix + "worker"
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
