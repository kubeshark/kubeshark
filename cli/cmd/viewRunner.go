package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/up9inc/mizu/cli/kubernetes"
	"github.com/up9inc/mizu/cli/mizu"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

func runMizuView() {
	kubernetesProvider := kubernetes.NewProvider("", "")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := kubernetesProvider.ClientSet.CoreV1().Services(mizu.ResourcesNamespace).Get(ctx, mizu.AggregatorPodName, metav1.GetOptions{})
	var statusError *k8serrors.StatusError
	if errors.As(err, &statusError) {
		if statusError.ErrStatus.Reason == metav1.StatusReasonNotFound {
			fmt.Printf("The %s service not found\n", mizu.AggregatorPodName)
			return
		}
		panic(err)
	}

	_, err = http.Get("http://localhost:8899/")
	if err == nil {
		fmt.Printf("Found a running service %s and open port 8899\n", mizu.AggregatorPodName)
		return
	}
	fmt.Printf("Found service %s, creating port forwarding to 8899\n", mizu.AggregatorPodName)
	portForwardApiPod(ctx, kubernetesProvider, cancel, &MizuTapOptions{GuiPort: 8899, MizuPodPort: 8899})
}
