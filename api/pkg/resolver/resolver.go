package resolver

import (
	"context"
	"errors"
	"fmt"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)
const (
	kubClientNullString = "None"
)

type Resolver struct {
	clientConfig  *restclient.Config
	clientSet *kubernetes.Clientset
	nameMap map[string]string
	serviceMap map[string]string
	isStarted bool
	errOut chan error
}

func (resolver *Resolver) Start(ctx context.Context) {
	if !resolver.isStarted {
		resolver.isStarted = true
		go resolver.infiniteErrorHandleRetryFunc(ctx, resolver.watchServices)
		go resolver.infiniteErrorHandleRetryFunc(ctx, resolver.watchEndpoints)
		go resolver.infiniteErrorHandleRetryFunc(ctx, resolver.watchPods)
	}
}

func (resolver *Resolver) Resolve(name string) string {
	resolvedName, isFound := resolver.nameMap[name]
	if !isFound {
		return ""
	}
	return resolvedName
}

func (resolver *Resolver) CheckIsServiceIP(address string) bool {
	_, isFound := resolver.serviceMap[address]
	return isFound
}

func (resolver *Resolver) watchPods(ctx context.Context) error {
	// empty namespace makes the client watch all namespaces
	watcher, err := resolver.clientSet.CoreV1().Pods("").Watch(ctx, metav1.ListOptions{Watch: true})
	if err != nil {
		return err
	}
	for {
		select {
			case event := <- watcher.ResultChan():
				if event.Object == nil {
					return errors.New("error in kubectl pod watch")
				}
				if event.Type == watch.Deleted {
					pod := event.Object.(*corev1.Pod)
					resolver.saveResolvedName(pod.Status.PodIP, "", event.Type)
				}
			case <- ctx.Done():
				watcher.Stop()
				return nil
		}
	}
}

func (resolver *Resolver) watchEndpoints(ctx context.Context) error {
	// empty namespace makes the client watch all namespaces
	watcher, err := resolver.clientSet.CoreV1().Endpoints("").Watch(ctx, metav1.ListOptions{Watch: true})
	if err != nil {
		return err
	}
	for {
		select {
			case event := <- watcher.ResultChan():
				if event.Object == nil {
					return errors.New("error in kubectl endpoint watch")
				}
				endpoint := event.Object.(*corev1.Endpoints)
				serviceHostname := fmt.Sprintf("%s.%s", endpoint.Name, endpoint.Namespace)
				if endpoint.Subsets != nil {
					for _, subset := range endpoint.Subsets {
						var ports []int32
						if subset.Ports != nil {
							for _, portMapping := range subset.Ports {
								if portMapping.Port > 0 {
									ports = append(ports, portMapping.Port)
								}
							}
						}
						if subset.Addresses != nil {
							for _, address := range subset.Addresses {
								resolver.saveResolvedName(address.IP, serviceHostname, event.Type)
								for _, port := range ports {
									ipWithPort := fmt.Sprintf("%s:%d", address.IP, port)
									resolver.saveResolvedName(ipWithPort, serviceHostname, event.Type)
								}
							}
						}

					}
				}
			case <- ctx.Done():
				watcher.Stop()
				return nil
		}
	}
}

func (resolver *Resolver) watchServices(ctx context.Context) error {
	// empty namespace makes the client watch all namespaces
	watcher, err := resolver.clientSet.CoreV1().Services("").Watch(ctx, metav1.ListOptions{Watch: true})
	if err != nil {
		return err
	}
	for {
		select {
		case event := <- watcher.ResultChan():
			if event.Object == nil {
				return errors.New("error in kubectl service watch")
			}

			service := event.Object.(*corev1.Service)
			serviceHostname := fmt.Sprintf("%s.%s", service.Name, service.Namespace)
			if service.Spec.ClusterIP != "" && service.Spec.ClusterIP != kubClientNullString {
				resolver.saveResolvedName(service.Spec.ClusterIP, serviceHostname, event.Type)
				resolver.saveServiceIP(service.Spec.ClusterIP, serviceHostname, event.Type)
			}
			if service.Status.LoadBalancer.Ingress != nil {
				for _, ingress := range service.Status.LoadBalancer.Ingress {
					resolver.saveResolvedName(ingress.IP, serviceHostname, event.Type)
				}
			}
		case <- ctx.Done():
			watcher.Stop()
			return nil
		}
	}
}

func (resolver *Resolver) saveResolvedName(key string, resolved string, eventType watch.EventType) {
	if eventType == watch.Deleted {
		delete(resolver.nameMap, key)
		// fmt.Printf("setting %s=nil\n", key)
	} else {
		resolver.nameMap[key] = resolved
		// fmt.Printf("setting %s=%s\n", key, resolved)
	}
}

func (resolver *Resolver) saveServiceIP(key string, resolved string, eventType watch.EventType) {
	if eventType == watch.Deleted {
		delete(resolver.serviceMap, key)
	} else {
		resolver.serviceMap[key] = resolved
	}
}

func (resolver *Resolver) infiniteErrorHandleRetryFunc(ctx context.Context, fun func(ctx context.Context) error) {
	for {
		err := fun(ctx)
		if err != nil {
			resolver.errOut <- err

			var statusError *k8serrors.StatusError
			if errors.As(err, &statusError) {
				if statusError.ErrStatus.Reason == metav1.StatusReasonForbidden {
					fmt.Printf("Resolver loop encountered permission error, aborting event listening - %v\n", err)
					return
				}
			}
		}
		if ctx.Err() != nil { // context was cancelled or errored
			return
		}
	}
}

