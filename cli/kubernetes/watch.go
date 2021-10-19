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

			for {
				watcher := kubernetesProvider.GetPodWatcher(ctx, targetNamespace)
				err, isContextCanceled := startWatchLoop(ctx, watcher, podFilter, addedChan, modifiedChan, removedChan)

				watcher.Stop()
				if isContextCanceled {
					break
				}

				if err != nil {
					errorChan <- errors.New("received too many unknown errors in k8s watch")
				} else {
					if !watchRestartDebouncer.IsOn() {
						watchRestartDebouncer.SetOn()
						logger.Log.Warning("detected a potential harmless watch timeout, retrying watch loop")
						time.Sleep(time.Second * 5)
						continue
					} else {
						errorChan <- errors.New("received too many unknown errors in k8s watch")
						break
					}
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

func startWatchLoop(ctx context.Context, watcher watch.Interface, podFilter *regexp.Regexp, addedChan chan *corev1.Pod, modifiedChan chan *corev1.Pod, removedChan chan *corev1.Pod) (err error, isContextCanceled bool) {
	for {
		select {
		case e := <-watcher.ResultChan():
			if e.Type == watch.Error {
				return apierrors.FromObject(e.Object), false
			}
			if e.Object == nil {
				return nil, false
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
			return nil, true
		}
	}
}