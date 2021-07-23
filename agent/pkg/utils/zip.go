package utils

import (
	"archive/zip"
	"bytes"
)

func ZipData(files map[string][]byte) *bytes.Buffer {
	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)
	// Create a new zip archive.
	zipWriter := zip.NewWriter(buf)
	defer func() { _ = zipWriter.Close() }()

	for fileName, fileBytes := range files {
		zipFile, _ := zipWriter.Create(fileName)
		_, _ = zipFile.Write(fileBytes)
	}
	return buf
}
