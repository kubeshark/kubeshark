package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"github.com/up9inc/mizu/shared/debounce"
	"github.com/up9inc/mizu/shared/logger"
	"sync"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/watch"
)

type EventFilterer interface {
	Filter(*watch.Event) (bool, error)
}

type WatchCreator interface {
	NewWatcher(ctx context.Context, namespace string) (watch.Interface, error)
}

func FilteredWatch(ctx context.Context, watcherCreator WatchCreator, targetNamespaces []string, filterer EventFilterer) (chan *watch.Event, chan *watch.Event, chan *watch.Event, chan error) {
	addedChan := make(chan *watch.Event)
	modifiedChan := make(chan *watch.Event)
	removedChan := make(chan *watch.Event)
	errorChan := make(chan error)

	var wg sync.WaitGroup

	for _, targetNamespace := range targetNamespaces {
		wg.Add(1)

		go func(targetNamespace string) {
			defer wg.Done()
			watchRestartDebouncer := debounce.NewDebouncer(1 * time.Minute, func() {})

			for {
				watcher, err := watcherCreator.NewWatcher(ctx, targetNamespace)
				if err != nil {
					errorChan <- fmt.Errorf("error in k8 watch: %v", err)
					break
				}

				err = startWatchLoop(ctx, watcher, filterer, addedChan, modifiedChan, removedChan) // blocking
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

func startWatchLoop(ctx context.Context, watcher watch.Interface, filterer EventFilterer, addedChan chan *watch.Event, modifiedChan chan *watch.Event, removedChan chan *watch.Event) error {
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

			if pass, err := filterer.Filter(&e); err != nil {
				return err
			} else if !pass {
				continue
			}

			switch e.Type {
			case watch.Added:
				addedChan <- &e
			case watch.Modified:
				modifiedChan <- &e
			case watch.Deleted:
				removedChan <- &e
			}
		case <-ctx.Done():
			return nil
		}
	}
}
