package check

import (
	"fmt"
	"log"

	"github.com/kubeshark/kubeshark/cli/kubernetes"
	"github.com/kubeshark/kubeshark/cli/semver"
	"github.com/kubeshark/kubeshark/cli/uiUtils"
)

func KubernetesVersion(kubernetesVersion *semver.SemVersion) bool {
	log.Printf("\nkubernetes-version\n--------------------")

	if err := kubernetes.ValidateKubernetesVersion(kubernetesVersion); err != nil {
		log.Printf("%v not running the minimum Kubernetes API version, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
		return false
	}

	log.Printf("%v is running the minimum Kubernetes API version", fmt.Sprintf(uiUtils.Green, "√"))
	return true
}
