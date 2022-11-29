package kubernetes

type K8sDeployManagerErrorReason string

const (
	DeployManagerWorkerUpdateError K8sDeployManagerErrorReason = "TAPPER_UPDATE_ERROR"
	DeployManagerPodWatchError     K8sDeployManagerErrorReason = "POD_WATCH_ERROR"
	DeployManagerPodListError      K8sDeployManagerErrorReason = "POD_LIST_ERROR"
)

type K8sDeployManagerError struct {
	OriginalError       error
	DeployManagerReason K8sDeployManagerErrorReason
}

// K8sDeployManagerError implements the Error interface.
func (e *K8sDeployManagerError) Error() string {
	return e.OriginalError.Error()
}

type ClusterBehindProxyError struct{}

// ClusterBehindProxyError implements the Error interface.
func (e *ClusterBehindProxyError) Error() string {
	return "Cluster is behind proxy"
}
