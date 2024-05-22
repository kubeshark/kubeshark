package fsUtils

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/misc"
	"github.com/rs/zerolog/log"
)

func DumpLogs(ctx context.Context, provider *kubernetes.Provider, filePath string, grep string) error {
	podExactRegex := regexp.MustCompile("^" + kubernetes.SELF_RESOURCES_PREFIX)
	pods, err := provider.ListAllPodsMatchingRegex(ctx, podExactRegex, []string{config.Config.Tap.Release.Namespace})
	if err != nil {
		return err
	}

	if len(pods) == 0 {
		return fmt.Errorf("No %s pods found in namespace %s", misc.Software, config.Config.Tap.Release.Namespace)
	}

	newZipFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer newZipFile.Close()
	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	for _, pod := range pods {
		for _, container := range pod.Spec.Containers {
			logs, err := provider.GetPodLogs(ctx, pod.Namespace, pod.Name, container.Name, grep)
			if err != nil {
				log.Error().Err(err).Msg("Failed to get logs!")
				continue
			} else {
				log.Debug().
					Int("length", len(logs)).
					Str("namespace", pod.Namespace).
					Str("pod", pod.Name).
					Str("container", container.Name).
					Msg("Successfully read log length.")
			}

			if err := AddStrToZip(zipWriter, logs, fmt.Sprintf("%s.%s.%s.log", pod.Namespace, pod.Name, container.Name)); err != nil {
				log.Error().Err(err).Msg("Failed write logs!")
			} else {
				log.Debug().
					Int("length", len(logs)).
					Str("namespace", pod.Namespace).
					Str("pod", pod.Name).
					Str("container", container.Name).
					Msg("Successfully added log length.")
			}
		}
	}

	events, err := provider.GetNamespaceEvents(ctx, config.Config.Tap.Release.Namespace)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get k8b events!")
	} else {
		log.Debug().Str("namespace", config.Config.Tap.Release.Namespace).Msg("Successfully read events.")
	}

	if err := AddStrToZip(zipWriter, events, fmt.Sprintf("%s_events.log", config.Config.Tap.Release.Namespace)); err != nil {
		log.Error().Err(err).Msg("Failed write logs!")
	} else {
		log.Debug().Str("namespace", config.Config.Tap.Release.Namespace).Msg("Successfully added events.")
	}

	if err := AddFileToZip(zipWriter, config.ConfigFilePath); err != nil {
		log.Error().Err(err).Msg("Failed write file!")
	} else {
		log.Debug().Str("file-path", config.ConfigFilePath).Msg("Successfully added file.")
	}

	log.Info().Str("path", filePath).Msg("You can find the ZIP file with all logs at:")
	return nil
}
