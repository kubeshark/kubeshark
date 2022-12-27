package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/docker"
	"github.com/kubeshark/kubeshark/internal/connect"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

func pullImages(ctx context.Context, cli *client.Client, imageFront string, imageHub string, imageWorker string) error {
	readerFront, err := cli.ImagePull(ctx, imageFront, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer readerFront.Close()
	_, err = io.Copy(os.Stdout, readerFront)
	if err != nil {
		return err
	}

	readerHub, err := cli.ImagePull(ctx, imageHub, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer readerHub.Close()
	_, err = io.Copy(os.Stdout, readerHub)
	if err != nil {
		return err
	}

	readerWorker, err := cli.ImagePull(ctx, imageWorker, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer readerWorker.Close()
	_, err = io.Copy(os.Stdout, readerWorker)
	if err != nil {
		return err
	}

	return nil
}

func createAndStartContainers(ctx context.Context, cli *client.Client, imageFront string, imageHub string, imageWorker string) (
	respFront container.ContainerCreateCreatedBody,
	respHub container.ContainerCreateCreatedBody,
	respWorker container.ContainerCreateCreatedBody,
	workerIPAddr string,
	err error,
) {
	hostIP := "0.0.0.0"

	hostConfigFront := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(fmt.Sprintf("%d/tcp", config.Config.Tap.Front.DstPort)): []nat.PortBinding{
				{
					HostIP:   hostIP,
					HostPort: fmt.Sprintf("%d", config.Config.Tap.Front.SrcPort),
				},
			},
		},
	}

	respFront, err = cli.ContainerCreate(ctx, &container.Config{
		Image: imageFront,
		Tty:   false,
	}, hostConfigFront, nil, nil, "kubeshark-front")
	if err != nil {
		return
	}

	if err = cli.ContainerStart(ctx, respFront.ID, types.ContainerStartOptions{}); err != nil {
		return
	}

	hostConfigHub := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(fmt.Sprintf("%d/tcp", config.Config.Tap.Hub.DstPort)): []nat.PortBinding{
				{
					HostIP:   hostIP,
					HostPort: fmt.Sprintf("%d", config.Config.Tap.Hub.SrcPort),
				},
			},
		},
	}

	cmdHub := []string{"-port", fmt.Sprintf("%d", config.Config.Tap.Hub.DstPort)}
	if config.DebugMode {
		cmdHub = append(cmdHub, fmt.Sprintf("-%s", config.DebugFlag))
	}

	respHub, err = cli.ContainerCreate(ctx, &container.Config{
		Image:        imageHub,
		Cmd:          cmdHub,
		Tty:          false,
		ExposedPorts: nat.PortSet{nat.Port(fmt.Sprintf("%d/tcp", config.Config.Tap.Hub.DstPort)): {}},
	}, hostConfigHub, nil, nil, "kubeshark-hub")
	if err != nil {
		return
	}

	if err = cli.ContainerStart(ctx, respHub.ID, types.ContainerStartOptions{}); err != nil {
		return
	}

	cmdWorker := []string{"-i", "any", "-port", fmt.Sprintf("%d", config.Config.Tap.Worker.DstPort)}
	if config.DebugMode {
		cmdWorker = append(cmdWorker, fmt.Sprintf("-%s", config.DebugFlag))
	}

	respWorker, err = cli.ContainerCreate(ctx, &container.Config{
		Image: imageWorker,
		Cmd:   cmdWorker,
		Tty:   false,
	}, nil, nil, nil, "kubeshark-worker")
	if err != nil {
		return
	}

	if err = cli.ContainerStart(ctx, respWorker.ID, types.ContainerStartOptions{}); err != nil {
		return
	}

	var containerWorker types.ContainerJSON
	containerWorker, err = cli.ContainerInspect(ctx, respWorker.ID)
	if err != nil {
		return
	}

	workerIPAddr = containerWorker.NetworkSettings.IPAddress

	return
}

func stopAndRemoveContainers(
	ctx context.Context,
	cli *client.Client,
	respFront container.ContainerCreateCreatedBody,
	respHub container.ContainerCreateCreatedBody,
	respWorker container.ContainerCreateCreatedBody,
) (err error) {
	err = cli.ContainerStop(ctx, respFront.ID, nil)
	if err != nil {
		return
	}
	err = cli.ContainerStop(ctx, respHub.ID, nil)
	if err != nil {
		return
	}
	err = cli.ContainerStop(ctx, respWorker.ID, nil)
	if err != nil {
		return
	}

	err = cli.ContainerRemove(ctx, respFront.ID, types.ContainerRemoveOptions{})
	if err != nil {
		return
	}
	err = cli.ContainerRemove(ctx, respHub.ID, types.ContainerRemoveOptions{})
	if err != nil {
		return
	}
	err = cli.ContainerRemove(ctx, respWorker.ID, types.ContainerRemoveOptions{})
	if err != nil {
		return
	}

	return
}

func pcap() {
	log.Info().Msg("Starting Docker containers...")

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
	defer cli.Close()

	imageFront := docker.GetFrontImage()
	imageHub := docker.GetHubImage()
	imageWorker := docker.GetWorkerImage()

	err = pullImages(ctx, cli, imageFront, imageHub, imageWorker)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	respFront, respHub, respWorker, workerIPAddr, err := createAndStartContainers(ctx, cli, imageFront, imageHub, imageWorker)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	workerPod := &v1.Pod{
		Spec: v1.PodSpec{
			NodeName: "docker",
		},
		Status: v1.PodStatus{
			PodIP: workerIPAddr,
			Phase: v1.PodRunning,
			ContainerStatuses: []v1.ContainerStatus{
				{
					Ready: true,
				},
			},
		},
	}

	connector = connect.NewConnector(kubernetes.GetLocalhostOnPort(config.Config.Tap.Hub.SrcPort), connect.DefaultRetries, connect.DefaultTimeout)
	connector.PostWorkerPodToHub(workerPod)

	log.Info().
		Str("url", kubernetes.GetLocalhostOnPort(config.Config.Tap.Hub.SrcPort)).
		Msg(fmt.Sprintf(utils.Green, "Hub is available at:"))

	url := kubernetes.GetLocalhostOnPort(config.Config.Tap.Front.SrcPort)
	log.Info().Str("url", url).Msg(fmt.Sprintf(utils.Green, "Kubeshark is available at:"))

	if !config.Config.HeadlessMode {
		utils.OpenBrowser(url)
	}

	ctxC, cancel := context.WithCancel(context.Background())
	defer cancel()
	utils.WaitForTermination(ctxC, cancel)

	err = stopAndRemoveContainers(ctx, cli, respFront, respHub, respWorker)
	if err != nil {
		log.Error().Err(err).Send()
	}
}