package source

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/up9inc/mizu/shared/logger"
	v1 "k8s.io/api/core/v1"
)

const envoyBinary = "/envoy"

var numberRegex = regexp.MustCompile("[0-9]+")

func discoverRelevantEnvoyPids(procfs string, pods []v1.Pod) ([]string, error) {
	result := make([]string, 0)

	pids, err := ioutil.ReadDir(procfs)

	if err != nil {
		return result, err
	}

	logger.Log.Infof("Starting envoy auto discoverer %v %v - scanning %v potential pids",
		procfs, pods, len(pids))

	for _, pid := range pids {
		if !pid.IsDir() {
			continue
		}

		if !numberRegex.MatchString(pid.Name()) {
			continue
		}

		if checkPid(procfs, pid.Name(), pods) {
			result = append(result, pid.Name())
		}
	}

	logger.Log.Infof("Found %v relevant envoy processes - %v", len(result), result)

	return result, nil
}

func checkPid(procfs string, pid string, pods []v1.Pod) bool {
	execLink := fmt.Sprintf("%v/%v/exe", procfs, pid)
	exec, err := os.Readlink(execLink)

	if err != nil {
		// Debug on purpose - it may happen due to many reasons and we only care
		//	for it during troubleshooting
		//
		logger.Log.Debugf("Unable to read link %v - %v\n", execLink, err)
		return false
	}

	if !strings.HasSuffix(exec, envoyBinary) {
		return false
	}

	environmentFile := fmt.Sprintf("%v/%v/environ", procfs, pid)
	podIp, err := readEnvironmentVariable(environmentFile, "INSTANCE_IP")

	if err != nil {
		return false
	}

	if podIp == "" {
		logger.Log.Debugf("Found an envoy process without INSTANCE_IP variable %v\n", pid)
		return false
	}

	logger.Log.Infof("Found envoy pid %v with cluster ip %v", pid, podIp)

	for _, pod := range pods {
		if pod.Status.PodIP == podIp {
			return true
		}
	}

	return false
}

func readEnvironmentVariable(file string, name string) (string, error) {
	bytes, err := ioutil.ReadFile(file)

	if err != nil {
		logger.Log.Warningf("Error reading environment file %v - %v", file, err)
		return "", err
	}

	envs := strings.Split(string(bytes), string([]byte{0}))

	for _, env := range envs {
		if !strings.Contains(env, "=") {
			continue
		}

		parts := strings.Split(env, "=")
		varName := parts[0]
		value := parts[1]

		if name == varName {
			return value, nil
		}
	}

	return "", nil
}
