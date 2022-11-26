package check

import (
	"fmt"
	"log"

	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/semver"
	"github.com/kubeshark/kubeshark/utils"
)

func KubernetesVersion(kubernetesVersion *semver.SemVersion) bool {
	log.Printf("\nkubernetes-version\n--------------------")

	if err := kubernetes.ValidateKubernetesVersion(kubernetesVersion); err != nil {
		log.Printf("%v not running the minimum Kubernetes API version, err: %v", fmt.Sprintf(utils.Red, "✗"), err)
		return false
	}

	log.Printf("%v is running the minimum Kubernetes API version", fmt.Sprintf(utils.Green, "√"))
	return true
}
