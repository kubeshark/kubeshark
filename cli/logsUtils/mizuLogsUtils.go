package logsUtils

import (
	"archive/zip"
	"context"
	"fmt"
	"github.com/up9inc/mizu/cli/kubernetes"
	"github.com/up9inc/mizu/cli/mizu"
	"io"
	"os"
	"path/filepath"
)

func DumpLogs(provider *kubernetes.Provider, ctx context.Context, filePath string) error {
	pods, err := provider.GetPods(ctx, mizu.ResourcesNamespace)
	if err != nil {
		return err
	}

	if len(pods) == 0 {
		return fmt.Errorf("no pods found in namespace %s", mizu.ResourcesNamespace)
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
			mizu.Log.Errorf("Failed to get logs, %v", err)
			continue
		} else {
			mizu.Log.Debugf("Successfully read log length %d for pod: %s.%s", len(logs), pod.Namespace, pod.Name)
		}
		if err := addLogsToZip(zipWriter, logs, fmt.Sprintf("%s.%s.log", pod.Namespace, pod.Name)); err != nil {
			mizu.Log.Errorf("Failed write logs, %v", err)
		} else {
			mizu.Log.Infof("Successfully added log length %d from pod: %s.%s", len(logs), pod.Namespace, pod.Name)
		}
	}
	if err := addFileToZip(zipWriter, mizu.GetConfigFilePath()); err != nil {
		mizu.Log.Errorf("Failed write file, %v", err)
	} else {
		mizu.Log.Infof("Successfully added file %s", mizu.GetConfigFilePath())
	}
	if err := addFileToZip(zipWriter, mizu.GetLogFilePath()); err != nil {
		mizu.Log.Errorf("Failed write file, %v", err)
	} else {
		mizu.Log.Infof("Successfully added file %s", mizu.GetLogFilePath())
	}
	mizu.Log.Infof("You can find the zip with all logs in %s\n", filePath)
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filename string) error {

	fileToZip, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s, %w", filename, err)
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file information %s, %w", filename, err)
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	header.Name = filepath.Base(filename)

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create header in zip for %s, %w", filename, err)
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

func addLogsToZip(writer *zip.Writer, logs string, fileName string) error {
	if zipFile, err := writer.Create(fileName); err != nil {
		return fmt.Errorf("couldn't create a log file inside zip for %s, %w", fileName, err)
	} else {
		if _, err = zipFile.Write([]byte(logs)); err != nil {
			return fmt.Errorf("couldn't write logs to zip file: %s, %w", fileName, err)
		}
	}
	return nil
}
