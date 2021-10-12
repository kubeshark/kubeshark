package kubernetes

import (
	"context"
	"errors"
	"regexp"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func FilteredWatch(ctx context.Context, kubernetesProvider *Provider, targetNamespaces []string, podFilter *regexp.Regexp) (chan *corev1.Pod, chan *corev1.Pod, chan *corev1.Pod, chan error) {
	addedChan := make(chan *corev1.Pod)
	modifiedChan := make(chan *corev1.Pod)
	removedChan := make(chan *corev1.Pod)
	errorChan := make(chan error)

	var wg sync.WaitGroup

	for _, targetNamespace := range targetNamespaces {
		wg.Add(1)

		go func(targetNamespace string) {
			defer wg.Done()
			watcher := kubernetesProvider.GetPodWatcher(ctx, targetNamespace)

			for {
				select {
				case e := <-watcher.ResultChan():
					if e.Object == nil {
						errorChan <- errors.New("kubernetes pod watch failed")
						return
					}

					pod, ok := e.Object.(*corev1.Pod)
					if !ok {
						continue
					}

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
					return
				}
			}
		}(targetNamespace)
	}

	go func() {
		<-ctx.Done()
		wg.Wait()
		close(addedChan)
		close(modifiedChan)
		close(removedChan)
		close(errorChan)
	}()

	return addedChan, modifiedChan, removedChan, errorChan
}
