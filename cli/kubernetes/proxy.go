package kubernetes

import (
	"fmt"
	"k8s.io/kubectl/pkg/proxy"
	"net"
	"net/http"
	"net/url"
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

	mizuProxiedUrl := GetMizuCollectorProxiesHostAndPath(mizuPort, mizuNamespace, mizuServiceName)
	proxyHandler, err := proxy.NewProxyHandler(k8sProxyApiPrefix, filter, &kubernetesProvider.clientConfig, time.Second * 2)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.Handle(k8sProxyApiPrefix, proxyHandler)
	//work around to make static resources available to the dashboard (all .svgs will not load without this)
	mux.Handle("/static/", getRerouteHttpHandler(proxyHandler, mizuProxiedUrl))

	//l, err := server.Listen("127.0.0.1", int(mizuPort))
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", "127.0.0.1", int(mizuPort)))
	if err != nil {
		return err
	}

	server := http.Server{
		Handler: mux,
	}
	return server.Serve(l)
}

func GetMizuCollectorProxiesHostAndPath(mizuPort uint16, mizuNamespace string, mizuServiceName string) string {
	return fmt.Sprintf("localhost:%d/api/v1/namespaces/%s/services/%s:80/proxy", mizuPort, mizuNamespace, mizuServiceName)
}

// rewrites requests so they end up reaching the mizu-collector k8s service via the k8s proxy handler
func getRerouteHttpHandler(proxyHandler http.Handler, mizuProxyUrl string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newUrl, _ := url.Parse(fmt.Sprintf("http://%s%s", mizuProxyUrl, r.URL.Path))
		r.URL = newUrl
		proxyHandler.ServeHTTP(w, r)
	})
}
