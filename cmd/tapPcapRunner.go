package cmd

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/docker"
	"github.com/kubeshark/kubeshark/internal/connect"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/misc"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

func logPullingImage(image string, reader io.ReadCloser) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		text := scanner.Text()
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(text), &data); err != nil {
			log.Error().Err(err).Send()
			continue
		}

		var id string
		if val, ok := data["id"]; ok {
			id = val.(string)
		}

		var status string
		if val, ok := data["status"]; ok {
			status = val.(string)
		}

		var progress string
		if val, ok := data["progress"]; ok {
			progress = val.(string)
		}

		e := log.Info()
		if image != "" {
			e = e.Str("image", image)
		}

		if progress != "" {
			e = e.Str("progress", progress)
		}

		e.Msg(fmt.Sprintf("[%-12s] %-18s", id, status))
	}
}

func pullImages(ctx context.Context, cli *client.Client, imageFront string, imageHub string, imageWorker string) error {
	log.Info().Msg("Pulling images...")
	readerFront, err := cli.ImagePull(ctx, imageFront, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer readerFront.Close()
	logPullingImage(imageFront, readerFront)

	readerHub, err := cli.ImagePull(ctx, imageHub, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer readerHub.Close()
	logPullingImage(imageHub, readerHub)

	readerWorker, err := cli.ImagePull(ctx, imageWorker, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer readerWorker.Close()
	logPullingImage(imageWorker, readerWorker)

	return nil
}

func cleanUpOldContainers(
	ctx context.Context,
	cli *client.Client,
	nameFront string,
	nameHub string,
	nameWorker string,
) error {
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return err
	}

	for _, container := range containers {
		f := fmt.Sprintf("/%s", nameFront)
		h := fmt.Sprintf("/%s", nameHub)
		w := fmt.Sprintf("/%s", nameWorker)
		if utils.Contains(container.Names, f) || utils.Contains(container.Names, h) || utils.Contains(container.Names, w) {
			err = cli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{Force: true})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func createAndStartContainers(
	ctx context.Context,
	cli *client.Client,
	imageFront string,
	imageHub string,
	imageWorker string,
	tarReader io.Reader,
) (
	respFront container.ContainerCreateCreatedBody,
	respHub container.ContainerCreateCreatedBody,
	respWorker container.ContainerCreateCreatedBody,
	workerIPAddr string,
	err error,
) {
	log.Info().Msg("Creating containers...")

	nameFront := fmt.Sprintf("%s-front", misc.Program)
	nameHub := fmt.Sprintf("%s-hub", misc.Program)
	nameWorker := fmt.Sprintf("%s-worker", misc.Program)

	err = cleanUpOldContainers(ctx, cli, nameFront, nameHub, nameWorker)
	if err != nil {
		return
	}

	hostIP := "0.0.0.0"

	hostConfigFront := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(fmt.Sprintf("%d/tcp", configStructs.ContainerPort)): []nat.PortBinding{
				{
					HostIP:   hostIP,
					HostPort: fmt.Sprintf("%d", config.Config.Tap.Proxy.Front.Port),
				},
			},
		},
	}

	respFront, err = cli.ContainerCreate(ctx, &container.Config{
		Image: imageFront,
		Tty:   false,
		Env: []string{
			"REACT_APP_DEFAULT_FILTER= ",
			"REACT_APP_HUB_HOST= ",
			fmt.Sprintf("REACT_APP_HUB_PORT=:%d", config.Config.Tap.Proxy.Hub.Port),
		},
	}, hostConfigFront, nil, nil, nameFront)
	if err != nil {
		return
	}

	hostConfigHub := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(fmt.Sprintf("%d/tcp", config.Config.Tap.Proxy.Hub.SrvPort)): []nat.PortBinding{
				{
					HostIP:   hostIP,
					HostPort: fmt.Sprintf("%d", config.Config.Tap.Proxy.Hub.Port),
				},
			},
		},
	}

	cmdHub := []string{"-port", fmt.Sprintf("%d", config.Config.Tap.Proxy.Hub.SrvPort)}
	if config.DebugMode {
		cmdHub = append(cmdHub, fmt.Sprintf("-%s", config.DebugFlag))
	}

	respHub, err = cli.ContainerCreate(ctx, &container.Config{
		Image:        imageHub,
		Cmd:          cmdHub,
		Tty:          false,
		ExposedPorts: nat.PortSet{nat.Port(fmt.Sprintf("%d/tcp", config.Config.Tap.Proxy.Hub.SrvPort)): {}},
	}, hostConfigHub, nil, nil, nameHub)
	if err != nil {
		return
	}

	cmdWorker := []string{"-f", "./import", "-port", fmt.Sprintf("%d", config.Config.Tap.Proxy.Worker.SrvPort)}
	if config.DebugMode {
		cmdWorker = append(cmdWorker, fmt.Sprintf("-%s", config.DebugFlag))
	}

	respWorker, err = cli.ContainerCreate(ctx, &container.Config{
		Image: imageWorker,
		Cmd:   cmdWorker,
		Tty:   false,
	}, nil, nil, nil, nameWorker)
	if err != nil {
		return
	}

	if err = cli.CopyToContainer(ctx, respWorker.ID, "/app/import", tarReader, types.CopyToContainerOptions{}); err != nil {
		return
	}

	log.Info().Msg("Starting containers...")

	if err = cli.ContainerStart(ctx, respFront.ID, types.ContainerStartOptions{}); err != nil {
		return
	}

	if err = cli.ContainerStart(ctx, respHub.ID, types.ContainerStartOptions{}); err != nil {
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
	log.Warn().Msg("Stopping containers...")
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

	log.Warn().Msg("Removing containers...")
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

func downloadTarFromS3(s3Url string) (tarPath string, err error) {
	u, err := url.Parse(s3Url)
	if err != nil {
		return
	}

	bucket := u.Host
	key := u.Path[1:]

	var cfg aws.Config
	cfg, err = awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		return
	}

	client := s3.NewFromConfig(cfg)

	var listObjectsOutput *s3.ListObjectsV2Output
	listObjectsOutput, err = client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return
	}

	if key == "" {
		var tempDirPath string
		tempDirPath, err = os.MkdirTemp(os.TempDir(), "kubeshark_*")
		if err != nil {
			return
		}

		for _, object := range listObjectsOutput.Contents {
			key = *object.Key
			fullPath := filepath.Join(tempDirPath, key)
			err = os.MkdirAll(filepath.Dir(fullPath), os.ModePerm)
			if err != nil {
				return
			}

			var file *os.File
			file, err = os.Create(fullPath)
			if err != nil {
				return
			}

			log.Info().Str("bucket", bucket).Str("key", key).Msg("Downloading from S3")

			downloader := manager.NewDownloader(client)
			_, err = downloader.Download(context.TODO(), file, &s3.GetObjectInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(key),
			})
			if err != nil {
				return
			}
		}

		tarPath, err = tarDirectory(tempDirPath)
	} else {
		var file *os.File
		file, err = os.CreateTemp(os.TempDir(), filepath.Base(key))
		if err != nil {
			return
		}

		log.Info().Str("bucket", bucket).Str("key", key).Msg("Downloading from S3")

		downloader := manager.NewDownloader(client)
		_, err = downloader.Download(context.TODO(), file, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			return
		}

		tarPath = file.Name()
	}

	return
}

func tarDirectory(dirPath string) (string, error) {
	tarPath := fmt.Sprintf("%s.tar.gz", dirPath)

	var file *os.File
	file, err := os.Create(tarPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			return err
		}

		header := &tar.Header{
			Name:    path[len(dirPath)+1:],
			Size:    stat.Size(),
			Mode:    int64(stat.Mode()),
			ModTime: stat.ModTime(),
		}

		err = tarWriter.WriteHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(tarWriter, file)
		if err != nil {
			return err
		}

		return nil
	}

	err = filepath.Walk(dirPath, walker)
	if err != nil {
		return "", err
	}

	return tarPath, nil
}

func pcap(tarPath string) error {
	if strings.HasPrefix(tarPath, "s3://") {
		var err error
		tarPath, err = downloadTarFromS3(tarPath)
		if err != nil {
			log.Error().Err(err).Msg("Failed downloading from S3")
			return err
		}
	}

	log.Info().Str("tar-path", tarPath).Msg("Openning")

	docker.SetRegistry(config.Config.Tap.Docker.Registry)
	docker.SetTag(config.Config.Tap.Docker.Tag)

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Error().Err(err).Send()
		return err
	}
	defer cli.Close()

	imageFront := docker.GetFrontImage()
	imageHub := docker.GetHubImage()
	imageWorker := docker.GetWorkerImage()

	err = pullImages(ctx, cli, imageFront, imageHub, imageWorker)
	if err != nil {
		log.Error().Err(err).Send()
		return err
	}

	tarFile, err := os.Open(tarPath)
	if err != nil {
		log.Error().Err(err).Send()
		return err
	}
	defer tarFile.Close()
	tarReader := bufio.NewReader(tarFile)

	respFront, respHub, respWorker, workerIPAddr, err := createAndStartContainers(
		ctx,
		cli,
		imageFront,
		imageHub,
		imageWorker,
		tarReader,
	)
	if err != nil {
		log.Error().Err(err).Send()
		return err
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

	connector = connect.NewConnector(kubernetes.GetProxyOnPort(config.Config.Tap.Proxy.Hub.Port), connect.DefaultRetries, connect.DefaultTimeout)
	connector.PostWorkerPodToHub(workerPod)

	// License
	if config.Config.License != "" {
		connector.PostLicense(config.Config.License)
	}

	log.Info().
		Str("url", kubernetes.GetProxyOnPort(config.Config.Tap.Proxy.Hub.Port)).
		Msg(fmt.Sprintf(utils.Green, "Hub is available at:"))

	url := kubernetes.GetProxyOnPort(config.Config.Tap.Proxy.Front.Port)
	log.Info().Str("url", url).Msg(fmt.Sprintf(utils.Green, fmt.Sprintf("%s is available at:", misc.Software)))

	if !config.Config.HeadlessMode {
		utils.OpenBrowser(url)
	}

	ctxC, cancel := context.WithCancel(context.Background())
	defer cancel()
	utils.WaitForTermination(ctxC, cancel)

	err = stopAndRemoveContainers(ctx, cli, respFront, respHub, respWorker)
	if err != nil {
		log.Error().Err(err).Send()
		return err
	}

	return nil
}
