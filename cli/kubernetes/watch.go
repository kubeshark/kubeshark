//   **Copied and modified from https://github.com/wercker/stern/blob/4fa46dd6987fca563d3ab42e61099658f4cade93/stern/watch.go**
//   Copyright 2016 Wercker Holding BV
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package kubernetes

import (
	"context"
	"fmt"
	"regexp"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// Target is a target to watch
type Target struct {
	Namespace string
	Pod       string
	Container string
}

// GetID returns the ID of the object
func (t *Target) GetID() string {
	return fmt.Sprintf("%s-%s-%s", t.Namespace, t.Pod, t.Container)
}

// FilteredWatch starts listening to Kubernetes events and emits modified
// containers/pods. The first result is targets added, the second is targets
// removed
func FilteredWatch(ctx context.Context, watcher watch.Interface, podFilter *regexp.Regexp, containerFilter *regexp.Regexp, containerExcludeFilter *regexp.Regexp) (chan *Target, chan *Target) {
	added := make(chan *Target)
	removed := make(chan *Target)
	go func() {
		for {
			select {
			case e := <-watcher.ResultChan():

				if e.Object == nil {
					// Closed because of error
					return
				}

				pod := e.Object.(*corev1.Pod)

				if !podFilter.MatchString(pod.Name) {
					continue
				}

				switch e.Type {
				case watch.Added:
					var statuses []corev1.ContainerStatus
					statuses = append(statuses, pod.Status.InitContainerStatuses...)
					statuses = append(statuses, pod.Status.ContainerStatuses...)

					added <- &Target{
						Namespace: pod.Namespace,
						Pod:       pod.Name,
						Container: "",
					}

					//for _, c := range statuses {
					//	if !containerFilter.MatchString(c.Name) {
					//		continue
					//	}
					//	if containerExcludeFilter != nil && containerExcludeFilter.MatchString(c.Name) {
					//		continue
					//	}
					//}
				case watch.Deleted:
					var containers []corev1.Container
					containers = append(containers, pod.Spec.Containers...)
					containers = append(containers, pod.Spec.InitContainers...)

					for _, c := range containers {
						//if !containerFilter.MatchString(c.Name) {
						//	continue
						//}
						//if containerExcludeFilter != nil && containerExcludeFilter.MatchString(c.Name) {
						//	continue
						//}

						removed <- &Target{
							Namespace: pod.Namespace,
							Pod:       pod.Name,
							Container: c.Name,
						}
					}
				}
			case <-ctx.Done():
				watcher.Stop()
				close(added)
				close(removed)
				return
			}
		}
	}()

	return added, removed
}
