package check

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/shared/kubernetes"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ImagePullInCluster(ctx context.Context, kubernetesProvider *kubernetes.Provider) bool {
	logger.Log.Infof("\nimage-pull-in-cluster\n--------------------")

	namespace := "default"
	podName := "mizu-test"

	defer func() {
		if err := kubernetesProvider.RemovePod(ctx, namespace, podName); err != nil {
			logger.Log.Errorf("%v error while removing test pod in cluster, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
		}
	}()

	if err := createImagePullInClusterPod(ctx, kubernetesProvider, namespace, podName); err != nil {
		logger.Log.Errorf("%v error while creating test pod in cluster, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
		return false
	}

	if err := checkImagePulled(ctx, kubernetesProvider, namespace, podName); err != nil {
		logger.Log.Errorf("%v cluster is not able to pull mizu containers from docker hub, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
		return false
	}

	logger.Log.Infof("%v cluster is able to pull mizu containers from docker hub", fmt.Sprintf(uiUtils.Green, "√"))
	return true
}

func checkImagePulled(ctx context.Context, kubernetesProvider *kubernetes.Provider, namespace string, podName string) error {
	podExactRegex := regexp.MustCompile(fmt.Sprintf("^%s$", podName))
	podWatchHelper := kubernetes.NewPodWatchHelper(kubernetesProvider, podExactRegex)
	eventChan, errorChan := kubernetes.FilteredWatch(ctx, podWatchHelper, []string{namespace}, podWatchHelper)

	timeAfter := time.After(30 * time.Second)

	for {
		select {
		case wEvent, ok := <-eventChan:
			if !ok {
				eventChan = nil
				continue
			}

			pod, err := wEvent.ToPod()
			if err != nil {
				return err
			}

			if pod.Status.Phase == core.PodRunning {
				return nil
			}
		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
				continue
			}

			return err
		case <-timeAfter:
			return fmt.Errorf("image not pulled in time")
		}
	}
}

func createImagePullInClusterPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, namespace string, podName string) error {
	var zero int64
	pod := &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:            "probe",
					Image:           "up9inc/busybox",
					ImagePullPolicy: "Always",
					Command:         []string{"cat"},
					Stdin:           true,
				},
			},
			TerminationGracePeriodSeconds: &zero,
		},
	}

	if _, err := kubernetesProvider.CreatePod(ctx, namespace, pod); err != nil {
		return err
	}

	return nil
}
