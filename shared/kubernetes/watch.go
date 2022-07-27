package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/shared/debounce"

	"k8s.io/apimachinery/pkg/watch"
)

type EventFilterer interface {
	Filter(*WatchEvent) (bool, error)
}

type WatchCreator interface {
	NewWatcher(ctx context.Context, namespace string) (watch.Interface, error)
}

func FilteredWatch(ctx context.Context, watcherCreator WatchCreator, targetNamespaces []string, filterer EventFilterer) (<-chan *WatchEvent, <-chan error) {
	eventChan := make(chan *WatchEvent)
	errorChan := make(chan error)

	var wg sync.WaitGroup

	for _, targetNamespace := range targetNamespaces {
		wg.Add(1)

		go func(targetNamespace string) {
			defer wg.Done()
			watchRestartDebouncer := debounce.NewDebouncer(1*time.Minute, func() {})

			for {
				watcher, err := watcherCreator.NewWatcher(ctx, targetNamespace)
				if err != nil {
					errorChan <- fmt.Errorf("error in k8s watch: %v", err)
					break
				}

				err = startWatchLoop(ctx, watcher, filterer, eventChan) // blocking
				watcher.Stop()

				select {
				case <-ctx.Done():
					return
				default:
					break
				}

				if err != nil {
					errorChan <- fmt.Errorf("error in k8s watch: %v", err)
					break
				} else {
					if !watchRestartDebouncer.IsOn() {
						if err := watchRestartDebouncer.SetOn(); err != nil {
							logger.Log.Error(err)
						}
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
		close(eventChan)
		close(errorChan)
	}()

	return eventChan, errorChan
}

func startWatchLoop(ctx context.Context, watcher watch.Interface, filterer EventFilterer, eventChan chan<- *WatchEvent) error {
	resultChan := watcher.ResultChan()
	for {
		select {
		case e, isChannelOpen := <-resultChan:
			if !isChannelOpen {
				return nil
			}

			wEvent := WatchEvent(e)

			if wEvent.Type == watch.Error {
				return wEvent.ToError()
			}

			if pass, err := filterer.Filter(&wEvent); err != nil {
				return err
			} else if !pass {
				continue
			}

			eventChan <- &wEvent
		case <-ctx.Done():
			return nil
		}
	}
}
