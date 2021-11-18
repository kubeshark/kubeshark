package kubernetes

import (
	"context"
	"regexp"

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
func (pwh *PodWatchHelper) Filter(wEvent *WatchEvent) (bool, error) {
	pod, err := wEvent.ToPod()
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
