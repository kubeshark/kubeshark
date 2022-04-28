package source

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/up9inc/mizu/logger"
	v1 "k8s.io/api/core/v1"
)

const linkerdBinary = "/linkerd2-proxy"

func discoverRelevantLinkerdPids(procfs string, pods []v1.Pod) ([]string, error) {
	result := make([]string, 0)

	pids, err := ioutil.ReadDir(procfs)

	if err != nil {
		return result, err
	}

	logger.Log.Infof("Starting linkerd auto discoverer %v %v - scanning %v potential pids",
		procfs, pods, len(pids))

	for _, pid := range pids {
		if !pid.IsDir() {
			continue
		}

		if !numberRegex.MatchString(pid.Name()) {
			continue
		}

		if checkLinkerdPid(procfs, pid.Name(), pods) {
			result = append(result, pid.Name())
		}
	}

	logger.Log.Infof("Found %v relevant linkerd processes - %v", len(result), result)

	return result, nil
}

func checkLinkerdPid(procfs string, pid string, pods []v1.Pod) bool {
	execLink := fmt.Sprintf("%v/%v/exe", procfs, pid)
	exec, err := os.Readlink(execLink)

	if err != nil {
		// Debug on purpose - it may happen due to many reasons and we only care
		//	for it during troubleshooting
		//
		logger.Log.Debugf("Unable to read link %v - %v\n", execLink, err)
		return false
	}

	if !strings.HasSuffix(exec, linkerdBinary) {
		return false
	}

	environmentFile := fmt.Sprintf("%v/%v/environ", procfs, pid)
	podName, err := getSingleValueFromEnvironmentVariableFile(environmentFile, "_pod_name")

	if err != nil {
		return false
	}

	if podName == "" {
		logger.Log.Debugf("Found a linkerd process without _pod_name variable %v\n", pid)
		return false
	}

	logger.Log.Infof("Found linkerd pid %v with pod name %v", pid, podName)

	for _, pod := range pods {
		if pod.Name == podName {
			return true
		}
	}

	return false
}
