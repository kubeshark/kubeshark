package kubernetes

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/up9inc/mizu/shared/logger"
	"k8s.io/kubectl/pkg/proxy"
)

const k8sProxyApiPrefix = "/"
const mizuServicePort = 80

func StartProxy(kubernetesProvider *Provider, proxyHost string, mizuPort uint16, mizuNamespace string, mizuServiceName string) error {
	logger.Log.Debugf("Starting proxy. namespace: [%v], service name: [%s], port: [%v]", mizuNamespace, mizuServiceName, mizuPort)
	filter := &proxy.FilterServer{
		AcceptPaths:   proxy.MakeRegexpArrayOrDie(proxy.DefaultPathAcceptRE),
		RejectPaths:   proxy.MakeRegexpArrayOrDie(proxy.DefaultPathRejectRE),
		AcceptHosts:   proxy.MakeRegexpArrayOrDie("^.*"),
		RejectMethods: proxy.MakeRegexpArrayOrDie(proxy.DefaultMethodRejectRE),
	}

	proxyHandler, err := proxy.NewProxyHandler(k8sProxyApiPrefix, filter, &kubernetesProvider.clientConfig, time.Second*2)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.Handle(k8sProxyApiPrefix, getRerouteHttpHandlerMizuAPI(proxyHandler, mizuNamespace, mizuServiceName))
	mux.Handle("/static/", getRerouteHttpHandlerMizuStatic(proxyHandler, mizuNamespace, mizuServiceName))

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", proxyHost, int(mizuPort)))
	if err != nil {
		return err
	}

	server := http.Server{
		Handler: mux,
	}

	return server.Serve(l)
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
