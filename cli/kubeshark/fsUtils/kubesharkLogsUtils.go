package fsUtils

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"path"
	"regexp"

	"github.com/kubeshark/kubeshark/cli/config"
	"github.com/kubeshark/kubeshark/cli/kubeshark"
	"github.com/kubeshark/kubeshark/logger"
	"github.com/kubeshark/kubeshark/shared/kubernetes"
)

func GetLogFilePath() string {
	return path.Join(kubeshark.GetKubesharkFolderPath(), "kubeshark_cli.log")
}

func DumpLogs(ctx context.Context, provider *kubernetes.Provider, filePath string) error {
	podExactRegex := regexp.MustCompile("^" + kubernetes.KubesharkResourcesPrefix)
	pods, err := provider.ListAllPodsMatchingRegex(ctx, podExactRegex, []string{config.Config.KubesharkResourcesNamespace})
	if err != nil {
		return err
	}

	if len(pods) == 0 {
		return fmt.Errorf("no kubeshark pods found in namespace %s", config.Config.KubesharkResourcesNamespace)
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
			logs, err := provider.GetPodLogs(ctx, pod.Namespace, pod.Name, container.Name)
			if err != nil {
				logger.Log.Errorf("Failed to get logs, %v", err)
				continue
			} else {
				logger.Log.Debugf("Successfully read log length %d for pod: %s.%s.%s", len(logs), pod.Namespace, pod.Name, container.Name)
			}

			if err := AddStrToZip(zipWriter, logs, fmt.Sprintf("%s.%s.%s.log", pod.Namespace, pod.Name, container.Name)); err != nil {
				logger.Log.Errorf("Failed write logs, %v", err)
			} else {
				logger.Log.Debugf("Successfully added log length %d from pod: %s.%s.%s", len(logs), pod.Namespace, pod.Name, container.Name)
			}
		}
	}

	events, err := provider.GetNamespaceEvents(ctx, config.Config.KubesharkResourcesNamespace)
	if err != nil {
		logger.Log.Debugf("Failed to get k8b events, %v", err)
	} else {
		logger.Log.Debugf("Successfully read events for k8b namespace: %s", config.Config.KubesharkResourcesNamespace)
	}

	if err := AddStrToZip(zipWriter, events, fmt.Sprintf("%s_events.log", config.Config.KubesharkResourcesNamespace)); err != nil {
		logger.Log.Debugf("Failed write logs, %v", err)
	} else {
		logger.Log.Debugf("Successfully added events for k8b namespace: %s", config.Config.KubesharkResourcesNamespace)
	}

	if err := AddFileToZip(zipWriter, config.Config.ConfigFilePath); err != nil {
		logger.Log.Debugf("Failed write file, %v", err)
	} else {
		logger.Log.Debugf("Successfully added file %s", config.Config.ConfigFilePath)
	}

	if err := AddFileToZip(zipWriter, GetLogFilePath()); err != nil {
		logger.Log.Debugf("Failed write file, %v", err)
	} else {
		logger.Log.Debugf("Successfully added file %s", GetLogFilePath())
	}

	logger.Log.Infof("You can find the zip file with all logs in %s", filePath)
	return nil
}
