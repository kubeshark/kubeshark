package kubernetes

import (
	"fmt"
	"regexp"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type PodWatchHelper struct {
	NameRegex *regexp.Regexp
}

// Implemets the EventFilterer Interface
func (pwh *PodWatchHelper) Filter(e *watch.Event) (bool, error) {
	pod, err := pwh.GetPodFromEvent(e);
	if err != nil {
		return false, nil
	}

	if !pwh.NameRegex.MatchString(pod.Name) {
		return false, nil
	}

	return true, nil
}

func (pwh *PodWatchHelper) GetPodFromEvent(e *watch.Event) (*core.Pod, error) {
	pod, ok := e.Object.(*core.Pod)
	if !ok {
		return nil, fmt.Errorf("Invalid object type on pod event stream")
	}

	return pod, nil
}


