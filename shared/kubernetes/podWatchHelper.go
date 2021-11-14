package kubernetes

import (
	"fmt"
	"regexp"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type PodWatchHelper struct {
	NameRegexFilter *regexp.Regexp
}

// Implemets the EventFilterer Interface
func (pwh *PodWatchHelper) Filter(e *watch.Event) (bool, error) {
	pod, err := pwh.GetPodFromEvent(e);
	if err != nil {
		return false, nil
	}

	if !pwh.NameRegexFilter.MatchString(pod.Name) {
		return false, nil
	}

	return true, nil
}

func (pwh *PodWatchHelper) GetPodFromEvent(e *watch.Event) (*corev1.Pod, error) {
	pod, ok := e.Object.(*corev1.Pod)
	if !ok {
		return nil, fmt.Errorf("Invalid object type on pod event stream")
	}

	return pod, nil
}
