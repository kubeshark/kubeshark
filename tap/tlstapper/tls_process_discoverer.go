package tlstapper

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/up9inc/mizu/shared/logger"
	v1 "k8s.io/api/core/v1"
)

var numberRegex = regexp.MustCompile("[0-9]+")

func UpdateTapTargets(tls *TlsTapper, pods *[]v1.Pod, procfs string) error {
	containerIds := buildContainerIdsMap(pods)
	containerPids, err := findContainerPids(procfs, containerIds)

	if err != nil {
		return err
	}

	for _, pid := range containerPids {
		if err := tls.AddPid(procfs, pid); err != nil {
			LogError(err)
		}
	}

	return nil
}

func findContainerPids(procfs string, containerIds map[string]bool) ([]uint32, error) {
	result := make([]uint32, 0)

	pids, err := ioutil.ReadDir(procfs)

	if err != nil {
		return result, err
	}

	logger.Log.Infof("Starting tls auto discoverer %v %v - scanning %v potential pids",
		procfs, containerIds, len(pids))

	for _, pid := range pids {
		if !pid.IsDir() {
			continue
		}

		if !numberRegex.MatchString(pid.Name()) {
			continue
		}

		cgroup, err := getProcessCgroup(procfs, pid.Name())

		if err != nil {
			continue
		}

		if _, ok := containerIds[cgroup]; !ok {
			continue
		}

		pidNumber, err := strconv.Atoi(pid.Name())

		if err != nil {
			continue
		}

		result = append(result, uint32(pidNumber))
	}

	return result, nil
}

func buildContainerIdsMap(pods *[]v1.Pod) map[string]bool {
	result := make(map[string]bool)

	for _, pod := range *pods {
		for _, container := range pod.Status.ContainerStatuses {
			url, err := url.Parse(container.ContainerID)

			if err != nil {
				logger.Log.Warningf("Expecting URL like container ID %v", container.ContainerID)
				continue
			}

			result[url.Host] = true
		}
	}

	return result
}

func getProcessCgroup(procfs string, pid string) (string, error) {
	filePath := fmt.Sprintf("%v/%v/cgroup", procfs, pid)

	bytes, err := ioutil.ReadFile(filePath)

	if err != nil {
		logger.Log.Warningf("Error reading cgroup file %v - %v", filePath, err)
		return "", err
	}
	
	lines := strings.Split(string(bytes), "\n")
	
	var cgrouppath string
	
	if len(lines) == 1 {
		parts := strings.Split(string(bytes), ":")
		cgrouppath = parts[len(parts)-1]
	} else {
		for _, line := range lines {
			if strings.Contains(line, ":pids:") {
				parts := strings.Split(line, ":")
				cgrouppath = parts[len(parts)-1]
			}
		}
	}
	
	basename := strings.TrimSpace(path.Base(cgrouppath))
	
	if strings.Contains(basename, ".") {
		return strings.TrimSuffix(basename, filepath.Ext(basename)), nil
	} else {
		return basename, nil
	}
}
