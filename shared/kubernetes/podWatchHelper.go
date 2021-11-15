package kubernetes

import (
	"context"
	"fmt"
	"regexp"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type PodWatchHelper struct {
	kubernetesProvider *Provider
	NameRegexFilter    *regexp.Regexp
}

func NewPodWatchHelper(kubernetesProvider *Provider, NameRegexFilter *regexp.Regexp) *PodWatchHelper {
	return &PodWatchHelper{
		kubernetesProvider: kubernetesProvider,
		NameRegexFilter: NameRegexFilter,
	}
}

// Implements the EventFilterer Interface
func (pwh *PodWatchHelper) Filter(e *watch.Event) (bool, error) {
	pod, err := pwh.GetPodFromEvent(e)
	if err != nil {
		return false, nil
	}

	if !pwh.NameRegexFilter.MatchString(pod.Name) {
		return false, nil
	}

	return true, nil
}

// Implements the WatchCreator Interface
func (pwh *PodWatchHelper) NewWatcher(ctx context.Context, namespace string) (watch.Interface, error) {
	watcher, err := pwh.kubernetesProvider.clientSet.CoreV1().Pods(namespace).Watch(ctx, metav1.ListOptions{Watch: true})
	if err != nil {
		return nil, err
	}

	return watcher, nil
}

func (pwh *PodWatchHelper) GetPodFromEvent(e *watch.Event) (*corev1.Pod, error) {
	pod, ok := e.Object.(*corev1.Pod)
	if !ok {
		return nil, fmt.Errorf("Invalid object type on pod event stream")
	}

	return pod, nil
}
