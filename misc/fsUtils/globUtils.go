package fsUtils

import (
	"fmt"
	"os"
	"path/filepath"
)

func RemoveFilesByExtension(dirPath string, ext string) error {
	files, err := filepath.Glob(filepath.Join(dirPath, fmt.Sprintf("/*.%s", ext)))
	if err != nil {
		return err
	}

	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}

	return nil
}
