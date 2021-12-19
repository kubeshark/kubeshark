package kubernetes

import (
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/watch"
)

const (
	EventAdded    watch.EventType = watch.Added
	EventModified watch.EventType = watch.Modified
	EventDeleted  watch.EventType = watch.Deleted
	EventBookmark watch.EventType = watch.Bookmark
	EventError    watch.EventType = watch.Error
)

type InvalidObjectType struct {
	RequestedType reflect.Type
}

// Implements the error interface
func (iot *InvalidObjectType) Error() string {
	return fmt.Sprintf("Cannot convert event to type %s", iot.RequestedType)
}

type WatchEvent watch.Event

func (we *WatchEvent) ToPod() (*corev1.Pod, error) {
	pod, ok := we.Object.(*corev1.Pod)
	if !ok {
		return nil, &InvalidObjectType{RequestedType: reflect.TypeOf(pod)}
	}

	return pod, nil
}

func (we *WatchEvent) ToEvent() (*eventsv1.Event, error) {
	event, ok := we.Object.(*eventsv1.Event)
	if !ok {
		return nil, &InvalidObjectType{RequestedType: reflect.TypeOf(event)}
	}

	return event, nil
}

func (we *WatchEvent) ToError() error {
	return apierrors.FromObject(we.Object)
}
