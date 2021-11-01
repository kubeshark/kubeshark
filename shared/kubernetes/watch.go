package kubernetes

import (
	"context"
	"errors"
	"fmt"
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
			watchRestartDebouncer := debounce.NewDebouncer(1 * time.Minute, func() {})

			for {
				watcher := kubernetesProvider.GetPodWatcher(ctx, targetNamespace)
				err := startWatchLoop(ctx, watcher, podFilter, addedChan, modifiedChan, removedChan) // blocking
				watcher.Stop()

				select {
				case <- ctx.Done():
					return
				default:
					break
				}

				if err != nil {
					errorChan <- fmt.Errorf("error in k8 watch: %v", err)
					break
				} else {
					if !watchRestartDebouncer.IsOn() {
						watchRestartDebouncer.SetOn()
						logger.Log.Debug("k8s watch channel closed, restarting watcher")
						time.Sleep(time.Second * 5)
						continue
					} else {
						errorChan <- errors.New("k8s watch unstable, closes frequently")
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

func startWatchLoop(ctx context.Context, watcher watch.Interface, podFilter *regexp.Regexp, addedChan chan *corev1.Pod, modifiedChan chan *corev1.Pod, removedChan chan *corev1.Pod) error {
	resultChan := watcher.ResultChan()
	for {
		select {
		case e, isChannelOpen := <-resultChan:
			if !isChannelOpen {
				return nil
			}

			if e.Type == watch.Error {
				return apierrors.FromObject(e.Object)
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
			return nil
		}
	}
}
