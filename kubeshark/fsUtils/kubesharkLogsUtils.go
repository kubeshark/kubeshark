package fsUtils

import (
	"archive/zip"
	"context"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/kubernetes"
)

func DumpLogs(ctx context.Context, provider *kubernetes.Provider, filePath string) error {
	podExactRegex := regexp.MustCompile("^" + kubernetes.KubesharkResourcesPrefix)
	pods, err := provider.ListAllPodsMatchingRegex(ctx, podExactRegex, []string{config.Config.ResourcesNamespace})
	if err != nil {
		return err
	}

	if len(pods) == 0 {
		return fmt.Errorf("no kubeshark pods found in namespace %s", config.Config.ResourcesNamespace)
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
				log.Printf("Failed to get logs, %v", err)
				continue
			} else {
				log.Printf("Successfully read log length %d for pod: %s.%s.%s", len(logs), pod.Namespace, pod.Name, container.Name)
			}

			if err := AddStrToZip(zipWriter, logs, fmt.Sprintf("%s.%s.%s.log", pod.Namespace, pod.Name, container.Name)); err != nil {
				log.Printf("Failed write logs, %v", err)
			} else {
				log.Printf("Successfully added log length %d from pod: %s.%s.%s", len(logs), pod.Namespace, pod.Name, container.Name)
			}
		}
	}

	events, err := provider.GetNamespaceEvents(ctx, config.Config.ResourcesNamespace)
	if err != nil {
		log.Printf("Failed to get k8b events, %v", err)
	} else {
		log.Printf("Successfully read events for k8b namespace: %s", config.Config.ResourcesNamespace)
	}

	if err := AddStrToZip(zipWriter, events, fmt.Sprintf("%s_events.log", config.Config.ResourcesNamespace)); err != nil {
		log.Printf("Failed write logs, %v", err)
	} else {
		log.Printf("Successfully added events for k8b namespace: %s", config.Config.ResourcesNamespace)
	}

	if err := AddFileToZip(zipWriter, config.Config.ConfigFilePath); err != nil {
		log.Printf("Failed write file, %v", err)
	} else {
		log.Printf("Successfully added file %s", config.Config.ConfigFilePath)
	}

	log.Printf("You can find the zip file with all logs in %s", filePath)
	return nil
}
