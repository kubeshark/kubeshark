package tlstapper

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/go-errors/errors"
	"github.com/up9inc/mizu/shared/logger"
)

func findSsllib(pid uint32) (string, error) {
	binary, err := os.Readlink(fmt.Sprintf("/proc/%v/exe", pid))

	if err != nil {
		return "", errors.Wrap(err, 0)
	}

	logger.Log.Infof("Binary file for %v = %v", pid, binary)

	if strings.HasSuffix(binary, "/node") {
		return findLibraryByPid(pid, binary)
	} else {
		return findLibraryByPid(pid, "libssl.so")
	}
}

func findLibraryByPid(pid uint32, libraryName string) (string, error) {
	file, err := os.Open(fmt.Sprintf("/proc/%v/maps", pid))

	if err != nil {
		return "", err
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())

		if len(parts) <= 5 {
			continue
		}

		filepath := parts[5]

		if !strings.Contains(filepath, libraryName) {
			continue
		}

		fullpath := fmt.Sprintf("/proc/%v/root/%v", pid, filepath)

		if _, err := os.Stat(fullpath); os.IsNotExist(err) {
			continue
		}

		return fullpath, nil
	}

	return "", errors.New(fmt.Sprintf("%v not found for PID %v", libraryName, pid))
}
