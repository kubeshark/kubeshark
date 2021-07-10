package kubernetes

import (
	"context"
	"fmt"
	"k8s.io/kubectl/pkg/proxy"
	"time"
)

func StartProxy(ctx context.Context, kubernetesProvider *Provider, mizuPort uint16, mizuNamespace string, mizuServiceName string) error {
	//o := cmdProxy.NewProxyOptions(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	filter := &proxy.FilterServer{
		AcceptPaths:   proxy.MakeRegexpArrayOrDie(proxy.DefaultPathAcceptRE),
		RejectPaths:   proxy.MakeRegexpArrayOrDie(proxy.DefaultPathRejectRE),
		AcceptHosts:   proxy.MakeRegexpArrayOrDie(proxy.DefaultHostAcceptRE),
		RejectMethods: proxy.MakeRegexpArrayOrDie(proxy.DefaultMethodRejectRE),
	}

	server, err := proxy.NewServer("", "/", "/static/", filter, &kubernetesProvider.clientConfig, time.Second * 1)

	l, err := server.Listen("127.0.0.1", int(mizuPort))
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("Closing connection due to context done")
				err := l.Close()
				if err != nil {
					fmt.Printf("Error stopping proxy network handler %v", err)
				}
				return
			}
		}
	}()
	fmt.Printf("Mizu is available at  http://localhost:%d/api/v1/namespaces/%s/services/%s:80/proxy\n", mizuPort, mizuNamespace, mizuServiceName)
	return server.ServeOnListener(l)
}
