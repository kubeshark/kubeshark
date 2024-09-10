package kubernetes

const (
	SELF_RESOURCES_PREFIX      = "kubeshark-"
	FrontPodName               = SELF_RESOURCES_PREFIX + "front"
	FrontServiceName           = FrontPodName
	HubPodName                 = SELF_RESOURCES_PREFIX + "hub"
	HubServiceName             = HubPodName
	K8sAllNamespaces           = ""
	MinKubernetesServerVersion = "1.16.0"
	AppLabelKey                = "app.kubeshark.co/app"
)
