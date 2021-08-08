package fsUtils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func AddFileToZip(zipWriter *zip.Writer, filename string) error {

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

func AddStrToZip(writer *zip.Writer, logs string, fileName string) error {
	if zipFile, err := writer.Create(fileName); err != nil {
		return fmt.Errorf("couldn't create a log file inside zip for %s, %w", fileName, err)
	} else {
		if _, err = zipFile.Write([]byte(logs)); err != nil {
			return fmt.Errorf("couldn't write logs to zip file: %s, %w", fileName, err)
		}
	}
	return nil
}
