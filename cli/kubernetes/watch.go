package kubernetes

import (
	"context"
	"errors"
	"regexp"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// FilteredWatch starts listening to Kubernetes events and emits modified
// containers/pods. The first result is targets added, the second is targets
// removed
func FilteredWatch(ctx context.Context, watcher watch.Interface, podFilter *regexp.Regexp) (chan *corev1.Pod, chan *corev1.Pod, chan *corev1.Pod, chan error) {
	addedChan := make(chan *corev1.Pod)
	modifiedChan := make(chan *corev1.Pod)
	removedChan := make(chan *corev1.Pod)
	errorChan := make(chan error)
	go func() {
		for {
			select {
			case e := <-watcher.ResultChan():

				if e.Object == nil {
					errorChan <- errors.New("kubernetes pod watch failed")
				}

				pod := e.Object.(*corev1.Pod)

				if !podFilter.MatchString(pod.Name) {
					continue
				}

				switch e.Type {
				case watch.Added:
					addedChan <- pod
				case watch.Modified:
					modifiedChan <- pod
				case watch.Deleted:
					removedChan <- pod
				}
			case <-ctx.Done():
				watcher.Stop()
				close(addedChan)
				close(modifiedChan)
				close(removedChan)
				close(errorChan)
				return
			}
		}
	}()

	return addedChan, modifiedChan, removedChan, errorChan
}
