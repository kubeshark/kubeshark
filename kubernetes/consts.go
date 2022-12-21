package kubernetes

const (
	KubesharkResourcesPrefix   = "kubeshark-"
	FrontPodName               = KubesharkResourcesPrefix + "front"
	FrontServiceName           = FrontPodName
	HubPodName                 = KubesharkResourcesPrefix + "hub"
	HubServiceName             = HubPodName
	ClusterRoleBindingName     = KubesharkResourcesPrefix + "cluster-role-binding"
	ClusterRoleName            = KubesharkResourcesPrefix + "cluster-role"
	K8sAllNamespaces           = ""
	RoleBindingName            = KubesharkResourcesPrefix + "role-binding"
	RoleName                   = KubesharkResourcesPrefix + "role"
	ServiceAccountName         = KubesharkResourcesPrefix + "service-account"
	WorkerDaemonSetName        = KubesharkResourcesPrefix + "worker-daemon-set"
	WorkerPodName              = KubesharkResourcesPrefix + "worker"
	ConfigMapName              = KubesharkResourcesPrefix + "config"
	MinKubernetesServerVersion = "1.16.0"
)

const (
	LabelPrefixApp      = "app.kubernetes.io/"
	LabelManagedBy      = LabelPrefixApp + "managed-by"
	LabelCreatedBy      = LabelPrefixApp + "created-by"
	LabelValueKubeshark = "kubeshark"
)
