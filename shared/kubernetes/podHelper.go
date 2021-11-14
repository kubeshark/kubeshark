package kubernetes

import (
	"fmt"
	"regexp"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type PodHelper struct {
	NameRegex *regexp.Regexp
}

// Implemets kubernetes.EventFilterer Interface
func (pef *PodHelper) Filter(e *watch.Event) (bool, error) {
	pod, err := pef.GetPodFromEvent(e);
	if err != nil {
		return false, nil
	}

	if !pef.NameRegex.MatchString(pod.Name) {
		return false, nil
	}

	return true, nil
}

func (pef *PodHelper) GetPodFromEvent(e *watch.Event) (*core.Pod, error) {
	pod, ok := e.Object.(*core.Pod)
	if !ok {
		return nil, fmt.Errorf("Invalid object type on pod event stream")
	}

	return pod, nil
}


