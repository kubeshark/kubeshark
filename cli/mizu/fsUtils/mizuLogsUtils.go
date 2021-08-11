package fsUtils

import (
	"archive/zip"
	"context"
	"fmt"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/kubernetes"
	"github.com/up9inc/mizu/cli/logger"
	"github.com/up9inc/mizu/cli/mizu"
	"os"
	"regexp"
)

func DumpLogs(provider *kubernetes.Provider, ctx context.Context, filePath string) error {
	podExactRegex := regexp.MustCompile("^" + mizu.MizuResourcesPrefix)
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
		logs, err := provider.GetPodLogs(pod.Namespace, pod.Name, ctx)
		if err != nil {
			logger.Log.Errorf("Failed to get logs, %v", err)
			continue
		} else {
			logger.Log.Debugf("Successfully read log length %d for pod: %s.%s", len(logs), pod.Namespace, pod.Name)
		}
		if err := AddStrToZip(zipWriter, logs, fmt.Sprintf("%s.%s.log", pod.Namespace, pod.Name)); err != nil {
			logger.Log.Errorf("Failed write logs, %v", err)
		} else {
			logger.Log.Infof("Successfully added log length %d from pod: %s.%s", len(logs), pod.Namespace, pod.Name)
		}
	}
	if err := AddFileToZip(zipWriter, config.GetConfigFilePath()); err != nil {
		logger.Log.Debugf("Failed write file, %v", err)
	} else {
		logger.Log.Infof("Successfully added file %s", config.GetConfigFilePath())
	}
	if err := AddFileToZip(zipWriter, logger.GetLogFilePath()); err != nil {
		logger.Log.Debugf("Failed write file, %v", err)
	} else {
		logger.Log.Infof("Successfully added file %s", logger.GetLogFilePath())
	}
	logger.Log.Infof("You can find the zip with all logs in %s\n", filePath)
	return nil
}
