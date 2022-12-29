package check

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/kubeshark/kubeshark/docker"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/misc"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/rs/zerolog/log"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ImagePullInCluster(ctx context.Context, kubernetesProvider *kubernetes.Provider) bool {
	log.Info().Str("procedure", "image-pull-in-cluster").Msg("Checking:")

	namespace := "default"
	podName := fmt.Sprintf("%s-test", misc.Program)

	defer func() {
		if err := kubernetesProvider.RemovePod(ctx, namespace, podName); err != nil {
			log.Error().
				Str("namespace", namespace).
				Str("pod", podName).
				Err(err).
				Msg("While removing test pod!")
		}
	}()

	if err := createImagePullInClusterPod(ctx, kubernetesProvider, namespace, podName); err != nil {
		log.Error().
			Str("namespace", namespace).
			Str("pod", podName).
			Err(err).
			Msg("While creating test pod!")
		return false
	}

	if err := checkImagePulled(ctx, kubernetesProvider, namespace, podName); err != nil {
		log.Printf("%v cluster is not able to pull %s containers from docker hub, err: %v", misc.Program, fmt.Sprintf(utils.Red, "✗"), err)
		log.Error().
			Str("namespace", namespace).
			Str("pod", podName).
			Err(err).
			Msg("Unable to pull images from Docker Hub!")
		return false
	}

	log.Info().
		Str("namespace", namespace).
		Str("pod", podName).
		Msg("Pulling images from Docker Hub is passed.")
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
	image := docker.GetWorkerImage()
	log.Info().Str("image", image).Msg("Testing image pull:")
	var zero int64
	pod := &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:            "probe",
					Image:           image,
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
