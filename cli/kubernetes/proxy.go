package kubernetes

import (
	"context"
	"fmt"
	"k8s.io/kubectl/pkg/proxy"
)

func StartProxy(ctx context.Context, kubernetesProvider *Provider, mizuPort uint16, mizuNamespace string, mizuServiceName string) error {
	server, err := proxy.NewServer("", "/", "", nil, &kubernetesProvider.clientConfig, 0)

	l, err := server.Listen("0.0.0.0", int(mizuPort))
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				err := l.Close()
				if err != nil {
					fmt.Printf("Error stopping proxy network handler %v", err)
				}
				return
			}
		}
	}()
	fmt.Printf("Mizu is available at  http://localhost:%d/api/v1/namespaces/%s/services/%s:80/proxy", mizuPort, mizuNamespace, mizuServiceName)
	return server.ServeOnListener(l)
}
