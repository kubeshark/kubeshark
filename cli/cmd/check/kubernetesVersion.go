package check

import (
	"fmt"

	"github.com/kubeshark/kubeshark/cli/uiUtils"
	"github.com/kubeshark/kubeshark/logger"
	"github.com/kubeshark/kubeshark/shared/kubernetes"
	"github.com/kubeshark/kubeshark/shared/semver"
)

func KubernetesVersion(kubernetesVersion *semver.SemVersion) bool {
	logger.Log.Infof("\nkubernetes-version\n--------------------")

	if err := kubernetes.ValidateKubernetesVersion(kubernetesVersion); err != nil {
		logger.Log.Errorf("%v not running the minimum Kubernetes API version, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
		return false
	}

	logger.Log.Infof("%v is running the minimum Kubernetes API version", fmt.Sprintf(uiUtils.Green, "√"))
	return true
}
