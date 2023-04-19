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
	PersistentVolumeName       = SelfResourcesPrefix + "persistent-volume"
	PersistentVolumeClaimName  = SelfResourcesPrefix + "persistent-volume-claim"
	PersistentVolumeHostPath   = "/app/data"
	MinKubernetesServerVersion = "1.16.0"
)

const (
	LabelManagedBy = SelfResourcesPrefix + "managed-by"
	LabelCreatedBy = SelfResourcesPrefix + "created-by"
)
