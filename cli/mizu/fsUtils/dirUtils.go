package fsUtils

import (
	"fmt"
	"os"
)

func EnsureDir(dirName string) error {
	err := os.Mkdir(dirName, 0700)
	if err == nil {
		return nil
	}
	if os.IsExist(err) {
		// check that the existing path is a directory
		info, err := os.Stat(dirName)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("path exists but is not a directory: %s", dirName)
		}
		return nil
	}
	return err
}
