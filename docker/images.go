package docker

import (
	"fmt"
	"strings"
)

const (
	hub    = "hub"
	worker = "worker"
	front  = "front"
)

var (
	registry             = "docker.io/kubeshark/"
	tag                  = "latest"
	NonCommunityRegistry = "public.ecr.aws/g4h6q0l5/kubeshark-"
)

func GetRegistry() string {
	return registry
}

func SetRegistry(value string) {
	if strings.HasPrefix(value, "docker.io/kubeshark") {
		registry = "docker.io/kubeshark/"
	} else {
		registry = value
	}
}

func GetTag() string {
	return tag
}

func SetTag(value string) {
	tag = value
}

func getImage(image string) string {
	return fmt.Sprintf("%s%s:%s", registry, image, tag)
}

func GetHubImage() string {
	return getImage(hub)
}

func GetWorkerImage() string {
	return getImage(worker)
}

func GetFrontImage() string {
	return getImage(front)
}
