package kubernetes

import (
	"context"
	"fmt"
	"regexp"

	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type EventWatchHelper struct {
	kubernetesProvider *Provider
	NameRegexFilter *regexp.Regexp
}

func NewEventWatchHelper(kubernetesProvider *Provider, NameRegexFilter *regexp.Regexp) *EventWatchHelper {
	return &EventWatchHelper{
		kubernetesProvider: kubernetesProvider,
		NameRegexFilter: NameRegexFilter,
	}
}

// Implements the EventFilterer Interface
func (pwh *EventWatchHelper) Filter(e *watch.Event) (bool, error) {
	event, err := pwh.GetEventFromEvent(e);
	if err != nil {
		return false, nil
	}

	if !pwh.NameRegexFilter.MatchString(event.Name) {
		return false, nil
	}

	return true, nil
}

// Implements the WatchCreator Interface
func (pwh *EventWatchHelper) NewWatcher(ctx context.Context, namespace string) (watch.Interface, error) {
	watcher, err := pwh.kubernetesProvider.clientSet.EventsV1().Events(namespace).Watch(ctx, metav1.ListOptions{Watch: true})
	if err != nil {
		return nil, err
	}

	return watcher, nil
}

func (pwh *EventWatchHelper) GetEventFromEvent(e *watch.Event) (*eventsv1.Event, error) {
	event, ok := e.Object.(*eventsv1.Event)
	if !ok {
		return nil, fmt.Errorf("Invalid object type on event event stream")
	}

	return event, nil
}
