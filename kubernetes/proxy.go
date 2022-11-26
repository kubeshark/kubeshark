package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

	"k8s.io/kubectl/pkg/proxy"
)

const k8sProxyApiPrefix = "/"
const kubesharkServicePort = 80

func StartProxy(kubernetesProvider *Provider, proxyHost string, srcPort uint16, dstPort uint16, kubesharkNamespace string, kubesharkServiceName string, cancel context.CancelFunc) (*http.Server, error) {
	log.Printf("Starting proxy - namespace: [%v], service name: [%s], port: [%d:%d]\n", kubesharkNamespace, kubesharkServiceName, srcPort, dstPort)
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
	mux.Handle(k8sProxyApiPrefix, getRerouteHttpHandlerKubesharkAPI(proxyHandler, kubesharkNamespace, kubesharkServiceName))
	mux.Handle("/static/", getRerouteHttpHandlerKubesharkStatic(proxyHandler, kubesharkNamespace, kubesharkServiceName))

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", proxyHost, int(srcPort)))
	if err != nil {
		return nil, err
	}

	server := &http.Server{
		Handler: mux,
	}

	go func() {
		if err := server.Serve(l); err != nil && err != http.ErrServerClosed {
			log.Printf("Error creating proxy, %v", err)
			cancel()
		}
	}()

	return server, nil
}

func getKubesharkHubProxiedHostAndPath(kubesharkNamespace string, kubesharkServiceName string) string {
	return fmt.Sprintf("/api/v1/namespaces/%s/services/%s:%d/proxy", kubesharkNamespace, kubesharkServiceName, kubesharkServicePort)
}

func GetLocalhostOnPort(port uint16) string {
	return fmt.Sprintf("http://localhost:%d", port)
}

func getRerouteHttpHandlerKubesharkAPI(proxyHandler http.Handler, kubesharkNamespace string, kubesharkServiceName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, x-session-token")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		proxiedPath := getKubesharkHubProxiedHostAndPath(kubesharkNamespace, kubesharkServiceName)

		//avoid redirecting several times
		if !strings.Contains(r.URL.Path, proxiedPath) {
			r.URL.Path = fmt.Sprintf("%s%s", getKubesharkHubProxiedHostAndPath(kubesharkNamespace, kubesharkServiceName), r.URL.Path)
		}
		proxyHandler.ServeHTTP(w, r)
	})
}

func getRerouteHttpHandlerKubesharkStatic(proxyHandler http.Handler, kubesharkNamespace string, kubesharkServiceName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.Replace(r.URL.Path, "/static/", fmt.Sprintf("%s/static/", getKubesharkHubProxiedHostAndPath(kubesharkNamespace, kubesharkServiceName)), 1)
		proxyHandler.ServeHTTP(w, r)
	})
}

func NewPortForward(kubernetesProvider *Provider, namespace string, podRegex *regexp.Regexp, srcPort uint16, dstPort uint16, ctx context.Context, cancel context.CancelFunc) (*portforward.PortForwarder, error) {
	pods, err := kubernetesProvider.ListAllRunningPodsMatchingRegex(ctx, podRegex, []string{namespace})
	if err != nil {
		return nil, err
	} else if len(pods) == 0 {
		return nil, fmt.Errorf("didn't find pod to port-forward")
	}

	podName := pods[0].Name

	log.Printf("Starting proxy using port-forward method. namespace: [%v], pod name: [%s], %d:%d", namespace, podName, srcPort, dstPort)

	dialer, err := getHttpDialer(kubernetesProvider, namespace, podName)
	if err != nil {
		return nil, err
	}

	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{}, 1)
	out, errOut := new(bytes.Buffer), new(bytes.Buffer)

	forwarder, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", srcPort, dstPort)}, stopChan, readyChan, out, errOut)
	if err != nil {
		return nil, err
	}

	go func() {
		if err = forwarder.ForwardPorts(); err != nil {
			log.Printf("kubernetes port-forwarding error: %v", err)
			cancel()
		}
	}()

	return forwarder, nil
}

func getHttpDialer(kubernetesProvider *Provider, namespace string, podName string) (httpstream.Dialer, error) {
	roundTripper, upgrader, err := spdy.RoundTripperFor(&kubernetesProvider.clientConfig)
	if err != nil {
		log.Printf("Error creating http dialer")
		return nil, err
	}

	clientConfigHostUrl, err := url.Parse(kubernetesProvider.clientConfig.Host)
	if err != nil {
		return nil, fmt.Errorf("Failed parsing client config host URL %s, error %w", kubernetesProvider.clientConfig.Host, err)
	}
	path := fmt.Sprintf("%s/api/v1/namespaces/%s/pods/%s/portforward", clientConfigHostUrl.Path, namespace, podName)

	serverURL := url.URL{Scheme: "https", Path: path, Host: clientConfigHostUrl.Host}
	log.Printf("Http dialer url %v", serverURL)

	return spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, &serverURL), nil
}
