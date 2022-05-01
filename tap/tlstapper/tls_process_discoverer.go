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

	"github.com/go-errors/errors"
	"github.com/up9inc/mizu/logger"
	v1 "k8s.io/api/core/v1"
)

var numberRegex = regexp.MustCompile("[0-9]+")

func UpdateTapTargets(tls *TlsTapper, pods *[]v1.Pod, procfs string) error {
	containerIds := buildContainerIdsMap(pods)
	containerPids, err := findContainerPids(procfs, containerIds)

	if err != nil {
		return err
	}

	tls.ClearPids()

	for pid, pod := range containerPids {
		if err := tls.AddPid(procfs, pid, pod.Namespace); err != nil {
			LogError(err)
		}
	}

	return nil
}

func findContainerPids(procfs string, containerIds map[string]v1.Pod) (map[uint32]v1.Pod, error) {
	result := make(map[uint32]v1.Pod)

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

		pod, ok := containerIds[cgroup]

		if !ok {
			continue
		}

		pidNumber, err := strconv.Atoi(pid.Name())

		if err != nil {
			continue
		}

		result[uint32(pidNumber)] = pod
	}

	return result, nil
}

func buildContainerIdsMap(pods *[]v1.Pod) map[string]v1.Pod {
	result := make(map[string]v1.Pod)

	for _, pod := range *pods {
		for _, container := range pod.Status.ContainerStatuses {
			url, err := url.Parse(container.ContainerID)

			if err != nil {
				logger.Log.Warningf("Expecting URL like container ID %v", container.ContainerID)
				continue
			}

			result[url.Host] = pod
		}
	}

	return result
}

func getProcessCgroup(procfs string, pid string) (string, error) {
	filePath := fmt.Sprintf("%s/%s/cgroup", procfs, pid)

	bytes, err := ioutil.ReadFile(filePath)

	if err != nil {
		logger.Log.Warningf("Error reading cgroup file %s - %v", filePath, err)
		return "", err
	}

	lines := strings.Split(string(bytes), "\n")
	cgrouppath := extractCgroup(lines)

	if cgrouppath == "" {
		return "", errors.Errorf("Cgroup path not found for %s, %s", pid, lines)
	}

	return normalizeCgroup(cgrouppath), nil
}

func extractCgroup(lines []string) string {
	if len(lines) == 1 {
		parts := strings.Split(lines[0], ":")
		return parts[len(parts)-1]
	} else {
		for _, line := range lines {
			if strings.Contains(line, ":pids:") {
				parts := strings.Split(line, ":")
				return parts[len(parts)-1]
			}
		}
	}

	return ""
}

// cgroup in the /proc/<pid>/cgroup may look something like
//
//  /system.slice/docker-<ID>.scope
//  /system.slice/containerd-<ID>.scope
//  /kubepods.slice/kubepods-burstable.slice/kubepods-burstable-pod3beae8e0_164d_4689_a087_efd902d8c2ab.slice/docker-<ID>.scope
//  /kubepods/besteffort/pod7709c1d5-447c-428f-bed9-8ddec35c93f4/<ID>
//
// This function extract the <ID> out of the cgroup path, the <ID> should match
//	the "Container ID:" field when running kubectl describe pod <POD>
//
func normalizeCgroup(cgrouppath string) string {
	basename := strings.TrimSpace(path.Base(cgrouppath))

	if strings.Contains(basename, "-") {
		basename = basename[strings.Index(basename, "-")+1:]
	}

	if strings.Contains(basename, ".") {
		return strings.TrimSuffix(basename, filepath.Ext(basename))
	} else {
		return basename
	}
}
