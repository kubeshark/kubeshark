package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/up9inc/mizu/shared"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

	"github.com/up9inc/mizu/logger"
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

	proxyHandler, err := proxy.NewProxyHandler(k8sProxyApiPrefix, filter, &kubernetesProvider.clientConfig, time.Second*2, false)
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

func NewPortForward(kubernetesProvider *Provider, namespace string, podRegex *regexp.Regexp, localPort uint16, ctx context.Context, cancel context.CancelFunc) (*portforward.PortForwarder, error) {
	pods, err := kubernetesProvider.ListAllRunningPodsMatchingRegex(ctx, podRegex, []string{namespace})
	if err != nil {
		return nil, err
	} else if len(pods) == 0 {
		return nil, fmt.Errorf("didn't find pod to port-forward")
	}

	podName := pods[0].Name

	logger.Log.Debugf("Starting proxy using port-forward method. namespace: [%v], pod name: [%s], port: [%v]", namespace, podName, localPort)

	dialer, err := getHttpDialer(kubernetesProvider, namespace, podName)
	if err != nil {
		return nil, err
	}

	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{}, 1)
	out, errOut := new(bytes.Buffer), new(bytes.Buffer)

	forwarder, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", localPort, shared.DefaultApiServerPort)}, stopChan, readyChan, out, errOut)
	if err != nil {
		return nil, err
	}

	go func() {
		if err = forwarder.ForwardPorts(); err != nil {
			logger.Log.Errorf("kubernetes port-forwarding error: %v", err)
			cancel()
		}
	}()

	return forwarder, nil
}

func getHttpDialer(kubernetesProvider *Provider, namespace string, podName string) (httpstream.Dialer, error) {
	roundTripper, upgrader, err := spdy.RoundTripperFor(&kubernetesProvider.clientConfig)
	if err != nil {
		logger.Log.Errorf("Error creating http dialer")
		return nil, err
	}

	clientConfigHostUrl, err := url.Parse(kubernetesProvider.clientConfig.Host)
	if err != nil {
		return nil, fmt.Errorf("Failed parsing client config host URL %s, error %w", kubernetesProvider.clientConfig.Host, err)
	}
	path := fmt.Sprintf("%s/api/v1/namespaces/%s/pods/%s/portforward", clientConfigHostUrl.Path, namespace, podName)

	serverURL := url.URL{Scheme: "https", Path: path, Host: clientConfigHostUrl.Host}
	logger.Log.Debugf("Http dialer url %v", serverURL)

	return spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, &serverURL), nil
}
