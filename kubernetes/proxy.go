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

	"github.com/kubeshark/kubeshark/config"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/kubectl/pkg/proxy"
)

const k8sProxyApiPrefix = "/"
const selfServicePort = 80

func StartProxy(kubernetesProvider *Provider, proxyHost string, srcPort uint16, selfNamespace string, selfServiceName string) (*http.Server, error) {
	log.Info().
		Str("proxy-host", proxyHost).
		Str("namespace", selfNamespace).
		Str("service", selfServiceName).
		Int("src-port", int(srcPort)).
		Msg("Starting proxy...")

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
	mux.Handle(k8sProxyApiPrefix, getRerouteHttpHandlerSelfAPI(proxyHandler, selfNamespace, selfServiceName))
	mux.Handle("/static/", getRerouteHttpHandlerSelfStatic(proxyHandler, selfNamespace, selfServiceName))

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", proxyHost, int(srcPort)))
	if err != nil {
		return nil, err
	}

	server := &http.Server{
		Handler: mux,
	}

	go func() {
		if err := server.Serve(l); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("While creating proxy!")
			return
		}
	}()

	return server, nil
}

func getSelfHubProxiedHostAndPath(selfNamespace string, selfServiceName string) string {
	return fmt.Sprintf("/api/v1/namespaces/%s/services/%s:%d/proxy", selfNamespace, selfServiceName, selfServicePort)
}

func GetProxyOnPort(port uint16) string {
	return fmt.Sprintf("http://%s:%d", config.Config.Tap.Proxy.Host, port)
}

func GetHubUrl() string {
	return fmt.Sprintf("%s/api", GetProxyOnPort(config.Config.Tap.Proxy.Front.Port))
}

func getRerouteHttpHandlerSelfAPI(proxyHandler http.Handler, selfNamespace string, selfServiceName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, x-session-token")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		proxiedPath := getSelfHubProxiedHostAndPath(selfNamespace, selfServiceName)

		//avoid redirecting several times
		if !strings.Contains(r.URL.Path, proxiedPath) {
			r.URL.Path = fmt.Sprintf("%s%s", getSelfHubProxiedHostAndPath(selfNamespace, selfServiceName), r.URL.Path)
		}
		proxyHandler.ServeHTTP(w, r)
	})
}

func getRerouteHttpHandlerSelfStatic(proxyHandler http.Handler, selfNamespace string, selfServiceName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.Replace(r.URL.Path, "/static/", fmt.Sprintf("%s/static/", getSelfHubProxiedHostAndPath(selfNamespace, selfServiceName)), 1)
		proxyHandler.ServeHTTP(w, r)
	})
}

func NewPortForward(kubernetesProvider *Provider, namespace string, podRegex *regexp.Regexp, srcPort uint16, dstPort uint16, ctx context.Context) (*portforward.PortForwarder, error) {
	pods, err := kubernetesProvider.ListPodsByAppLabel(ctx, namespace, map[string]string{AppLabelKey: "front"})
	if err != nil {
		return nil, err
	} else if len(pods) == 0 {
		return nil, fmt.Errorf("didn't find pod to port-forward")
	}

	podName := pods[0].Name

	log.Info().
		Str("namespace", namespace).
		Str("pod", podName).
		Int("src-port", int(srcPort)).
		Int("dst-port", int(dstPort)).
		Msg("Starting proxy using port-forward method...")

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
			log.Error().Err(err).Msg("While Kubernetes port-forwarding!")
			return
		}
	}()

	return forwarder, nil
}

func getHttpDialer(kubernetesProvider *Provider, namespace string, podName string) (httpstream.Dialer, error) {
	roundTripper, upgrader, err := spdy.RoundTripperFor(&kubernetesProvider.clientConfig)
	if err != nil {
		log.Error().Err(err).Msg("While creating HTTP dialer!")
		return nil, err
	}

	clientConfigHostUrl, err := url.Parse(kubernetesProvider.clientConfig.Host)
	if err != nil {
		return nil, fmt.Errorf("Failed parsing client config host URL %s, error %w", kubernetesProvider.clientConfig.Host, err)
	}
	path := fmt.Sprintf("%s/api/v1/namespaces/%s/pods/%s/portforward", clientConfigHostUrl.Path, namespace, podName)

	serverURL := url.URL{Scheme: "https", Path: path, Host: clientConfigHostUrl.Host}
	log.Debug().
		Str("url", serverURL.String()).
		Msg("HTTP dialer URL:")

	return spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, &serverURL), nil
}
