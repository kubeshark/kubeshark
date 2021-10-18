package kubernetes

import (
	"context"
	"errors"
	"github.com/up9inc/mizu/shared/debounce"
	"github.com/up9inc/mizu/shared/logger"
	"regexp"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
			watchRestartDebouncer := debounce.NewDebouncer(5 * time.Second, func() {})
			shouldStop := false

			for !shouldStop {
				watcher := kubernetesProvider.GetPodWatcher(ctx, targetNamespace)
				func() {
					for !shouldStop {
						select {
						case e := <-watcher.ResultChan():
							if e.Type == watch.Error {
								errorChan <- apierrors.FromObject(e.Object)
								shouldStop = true
								return
							}
							if e.Object == nil {
								if !watchRestartDebouncer.IsOn() {
									logger.Log.Debug("detected a potential harmless watch timeout, retrying watch loop")
									return
								} else {
									errorChan <- errors.New("received too many unknown errors in k8s watch")
									shouldStop = true
									return
								}

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
							shouldStop = true
							return
						}
					}
				}()
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
