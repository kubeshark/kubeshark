package docker

import "fmt"

const (
	hub    = "hub"
	worker = "worker"
	front  = "front"
)

var (
	registry = "docker.io/kubeshark/"
	tag      = "latest"
)

func GetRegistry() string {
	return registry
}

func SetRegistry(value string) {
	registry = value
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
