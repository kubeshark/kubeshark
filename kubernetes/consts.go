package kubernetes

const (
	SelfResourcesPrefix        = "kubeshark-"
	FrontPodName               = SelfResourcesPrefix + "front"
	FrontServiceName           = FrontPodName
	HubPodName                 = SelfResourcesPrefix + "hub"
	HubServiceName             = HubPodName
	ClusterRoleBindingName     = SelfResourcesPrefix + "cluster-role-binding"
	ClusterRoleName            = SelfResourcesPrefix + "cluster-role"
	K8sAllNamespaces           = ""
	RoleBindingName            = SelfResourcesPrefix + "role-binding"
	RoleName                   = SelfResourcesPrefix + "role"
	ServiceAccountName         = SelfResourcesPrefix + "service-account"
	WorkerDaemonSetName        = SelfResourcesPrefix + "worker-daemon-set"
	WorkerPodName              = SelfResourcesPrefix + "worker"
	ConfigMapName              = SelfResourcesPrefix + "config"
	MinKubernetesServerVersion = "1.16.0"
)

const (
	LabelPrefixApp = "app.kubernetes.io/"
	LabelManagedBy = LabelPrefixApp + "managed-by"
	LabelCreatedBy = LabelPrefixApp + "created-by"
)
