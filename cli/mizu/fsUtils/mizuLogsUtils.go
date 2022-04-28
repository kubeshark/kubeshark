package fsUtils

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"path"
	"regexp"

	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/shared/kubernetes"
)

func GetLogFilePath() string {
	return path.Join(mizu.GetMizuFolderPath(), "mizu_cli.log")
}

func DumpLogs(ctx context.Context, provider *kubernetes.Provider, filePath string) error {
	podExactRegex := regexp.MustCompile("^" + kubernetes.MizuResourcesPrefix)
	pods, err := provider.ListAllPodsMatchingRegex(ctx, podExactRegex, []string{config.Config.MizuResourcesNamespace})
	if err != nil {
		return err
	}

	if len(pods) == 0 {
		return fmt.Errorf("no mizu pods found in namespace %s", config.Config.MizuResourcesNamespace)
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

	events, err := provider.GetNamespaceEvents(ctx, config.Config.MizuResourcesNamespace)
	if err != nil {
		logger.Log.Debugf("Failed to get k8b events, %v", err)
	} else {
		logger.Log.Debugf("Successfully read events for k8b namespace: %s", config.Config.MizuResourcesNamespace)
	}

	if err := AddStrToZip(zipWriter, events, fmt.Sprintf("%s_events.log", config.Config.MizuResourcesNamespace)); err != nil {
		logger.Log.Debugf("Failed write logs, %v", err)
	} else {
		logger.Log.Debugf("Successfully added events for k8b namespace: %s", config.Config.MizuResourcesNamespace)
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
