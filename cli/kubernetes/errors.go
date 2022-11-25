package kubernetes

type K8sTapManagerErrorReason string

const (
	TapManagerTapperUpdateError K8sTapManagerErrorReason = "TAPPER_UPDATE_ERROR"
	TapManagerPodWatchError     K8sTapManagerErrorReason = "POD_WATCH_ERROR"
	TapManagerPodListError      K8sTapManagerErrorReason = "POD_LIST_ERROR"
)

type K8sTapManagerError struct {
	OriginalError    error
	TapManagerReason K8sTapManagerErrorReason
}

// K8sTapManagerError implements the Error interface.
func (e *K8sTapManagerError) Error() string {
	return e.OriginalError.Error()
}

type ClusterBehindProxyError struct{}

// ClusterBehindProxyError implements the Error interface.
func (e *ClusterBehindProxyError) Error() string {
	return "Cluster is behind proxy"
}
