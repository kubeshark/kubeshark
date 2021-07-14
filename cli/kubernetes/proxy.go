package kubernetes

import (
	"fmt"
	"k8s.io/kubectl/pkg/proxy"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const k8sProxyApiPrefix = "/"

func StartProxy(kubernetesProvider *Provider, mizuPort uint16, mizuNamespace string, mizuServiceName string) error {
	filter := &proxy.FilterServer{
		AcceptPaths:   proxy.MakeRegexpArrayOrDie(proxy.DefaultPathAcceptRE),
		RejectPaths:   proxy.MakeRegexpArrayOrDie(proxy.DefaultPathRejectRE),
		AcceptHosts:   proxy.MakeRegexpArrayOrDie(proxy.DefaultHostAcceptRE),
		RejectMethods: proxy.MakeRegexpArrayOrDie(proxy.DefaultMethodRejectRE),
	}

	proxyHandler, err := proxy.NewProxyHandler(k8sProxyApiPrefix, filter, &kubernetesProvider.clientConfig, time.Second * 2)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.Handle(k8sProxyApiPrefix, proxyHandler)
	mux.Handle("/static/", getRerouteHttpHandlerMizuAPI(proxyHandler, getMizuCollectorProxiedHostAndPath(mizuPort, mizuNamespace, mizuServiceName)))
	mux.Handle("/mizu/", getRerouteHttpHandlerMizuAPI(proxyHandler, getMizuCollectorProxiedHostAndPath(mizuPort, mizuNamespace, mizuServiceName)))


	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", "127.0.0.1", int(mizuPort)))
	if err != nil {
		return err
	}

	server := http.Server{
		Handler: mux,
	}
	return server.Serve(l)
}

const mizuServicePort = 80

func getMizuCollectorProxiedHostAndPath(mizuPort uint16, mizuNamespace string, mizuServiceName string) string {
	return fmt.Sprintf("localhost:%d/api/v1/namespaces/%s/services/%s:%v/proxy", mizuPort, mizuNamespace, mizuServiceName, mizuServicePort)
}

func GetMizuCollectorProxiedHostAndPath(mizuPort uint16) string {
	return fmt.Sprintf("localhost:%d/mizu", mizuPort)
}


// rewrites requests so they end up reaching the mizu-collector k8s service via the k8s proxy handler
func getRerouteHttpHandlerMizuAPI(proxyHandler http.Handler, mizuProxyUrl string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newPath := strings.Replace(r.URL.Path, "/mizu/", "/", 1)
		newUrl, _ := url.Parse(fmt.Sprintf("http://%s%s", mizuProxyUrl, newPath))
		r.URL = newUrl
		proxyHandler.ServeHTTP(w, r)
	})
}
