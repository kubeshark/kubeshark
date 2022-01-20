package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"github.com/up9inc/mizu/shared"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/up9inc/mizu/shared/logger"
	"k8s.io/kubectl/pkg/proxy"
)

const k8sProxyApiPrefix = "/"
const mizuServicePort = 80

func StartProxy(kubernetesProvider *Provider, proxyHost string, mizuPort uint16, mizuNamespace string, mizuServiceName string, cancel context.CancelFunc) (*http.Server, error) {
	logger.Log.Debugf("Starting proxy using proxy method. namespace: [%v], service name: [%s], port: [%v]", mizuNamespace, mizuServiceName, mizuPort)
	filter := &proxy.FilterServer{
		AcceptPaths:   proxy.MakeRegexpArrayOrDie(proxy.DefaultPathAcceptRE),
		RejectPaths:   proxy.MakeRegexpArrayOrDie(proxy.DefaultPathRejectRE),
		AcceptHosts:   proxy.MakeRegexpArrayOrDie("^.*"),
		RejectMethods: proxy.MakeRegexpArrayOrDie(proxy.DefaultMethodRejectRE),
	}

	proxyHandler, err := proxy.NewProxyHandler(k8sProxyApiPrefix, filter, &kubernetesProvider.clientConfig, time.Second*2)
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.Handle(k8sProxyApiPrefix, getRerouteHttpHandlerMizuAPI(proxyHandler, mizuNamespace, mizuServiceName))
	mux.Handle("/static/", getRerouteHttpHandlerMizuStatic(proxyHandler, mizuNamespace, mizuServiceName))

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", proxyHost, int(mizuPort)))
	if err != nil {
		return nil, err
	}

	server := &http.Server{
		Handler: mux,
	}

	go func() {
		if err := server.Serve(l); err != nil && err != http.ErrServerClosed {
			logger.Log.Errorf("Error creating proxy, %v", err)
			cancel()
		}
	}()

	return server, nil
}


func getMizuApiServerProxiedHostAndPath(mizuNamespace string, mizuServiceName string) string {
	return fmt.Sprintf("/api/v1/namespaces/%s/services/%s:%d/proxy", mizuNamespace, mizuServiceName, mizuServicePort)
}

func GetMizuApiServerProxiedHostAndPath(mizuPort uint16) string {
	return fmt.Sprintf("localhost:%d", mizuPort)
}

func getRerouteHttpHandlerMizuAPI(proxyHandler http.Handler, mizuNamespace string, mizuServiceName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxiedPath := getMizuApiServerProxiedHostAndPath(mizuNamespace, mizuServiceName)

		//avoid redirecting several times
		if !strings.Contains(r.URL.Path, proxiedPath) {
			r.URL.Path = fmt.Sprintf("%s%s", getMizuApiServerProxiedHostAndPath(mizuNamespace, mizuServiceName), r.URL.Path)
		}
		proxyHandler.ServeHTTP(w, r)
	})
}

func getRerouteHttpHandlerMizuStatic(proxyHandler http.Handler, mizuNamespace string, mizuServiceName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.Replace(r.URL.Path, "/static/", fmt.Sprintf("%s/static/", getMizuApiServerProxiedHostAndPath(mizuNamespace, mizuServiceName)), 1)
		proxyHandler.ServeHTTP(w, r)
	})
}

func NewPortForward(kubernetesProvider *Provider, namespace string, podName string, localPort uint16, cancel context.CancelFunc) error {
	logger.Log.Debugf("Starting proxy using port-forward method. namespace: [%v], service name: [%s], port: [%v]", namespace, podName, localPort)

	dialer, err := getHttpDialer(kubernetesProvider, namespace, podName)
	if err != nil {
		return err
	}

	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{}, 1)
	out, errOut := new(bytes.Buffer), new(bytes.Buffer)

	forwarder, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", localPort, shared.DefaultApiServerPort)}, stopChan, readyChan, out, errOut)
	if err != nil {
		return err
	}

	go func() {
		if err = forwarder.ForwardPorts(); err != nil {
			logger.Log.Errorf("kubernetes port-forwarding error: %v", err)
			cancel()
		}
	}()

	return nil
}

func getHttpDialer(kubernetesProvider *Provider, namespace string, podName string) (httpstream.Dialer, error) {
	roundTripper, upgrader, err := spdy.RoundTripperFor(&kubernetesProvider.clientConfig)
	if err != nil {
		logger.Log.Errorf("Error creating http dialer")
		return nil, err
	}
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", namespace, podName)
	hostIP := strings.TrimLeft(kubernetesProvider.clientConfig.Host, "htps:/") // no need specify "t" twice
	serverURL := url.URL{Scheme: "https", Path: path, Host: hostIP}

	return spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, &serverURL), nil
}
