package oas

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/up9inc/mizu/shared/logger"
	"io"
	"io/ioutil"
	"mizuserver/pkg/har"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func getFiles(baseDir string) (result []string, err error) {
	result = make([]string, 0, 0)
	logger.Log.Infof("Reading files from tree: %s", baseDir)

	// https://yourbasic.org/golang/list-files-in-directory/
	err = filepath.Walk(baseDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			ext := strings.ToLower(filepath.Ext(path))
			if !info.IsDir() && (ext == ".har" || ext == ".ldjson") {
				result = append(result, path)
			}

			return nil
		})

	sort.SliceStable(result, func(i, j int) bool {
		return fileSize(result[i]) < fileSize(result[j])
	})

	logger.Log.Infof("Got files: %d", len(result))
	return result, err
}

func fileSize(fname string) int64 {
	fi, err := os.Stat(fname)
	if err != nil {
		panic(err)
	}

	return fi.Size()
}

func feedEntries(fromFiles []string) (count int, err error) {
	badFiles := make([]string, 0)
	cnt := 0
	for _, file := range fromFiles {
		logger.Log.Info("Processing file: " + file)
		ext := strings.ToLower(filepath.Ext(file))
		eCnt := 0
		switch ext {
		case ".har":
			eCnt, err = feedFromHAR(file)
			if err != nil {
				logger.Log.Warning("Failed processing file: " + err.Error())
				badFiles = append(badFiles, file)
				continue
			}
		case ".ldjson":
			eCnt, err = feedFromLDJSON(file)
			if err != nil {
				logger.Log.Warning("Failed processing file: " + err.Error())
				badFiles = append(badFiles, file)
				continue
			}
		default:
			return 0, errors.New("Unsupported file extension: " + ext)
		}
		cnt += eCnt
	}

	for _, f := range badFiles {
		logger.Log.Infof("Bad file: %s", f)
	}

	return cnt, nil
}

func feedFromHAR(file string) (int, error) {
	fd, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	defer fd.Close()

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return 0, err
	}

	var harDoc har.HAR
	err = json.Unmarshal(data, &harDoc)
	if err != nil {
		return 0, err
	}

	cnt := 0
	for _, entry := range harDoc.Log.Entries {
		cnt += 1
		feedEntry(&entry)
	}

	return cnt, nil
}

func feedEntry(entry *har.Entry) {
	if entry.Response.Status == 302 {
		logger.Log.Debugf("Dropped traffic entry due to permanent redirect status: %s", entry.StartedDateTime)
	}
	// GetOasGeneratorInstance().PushEntry(entry)
	GetOasGeneratorInstance().entriesChan <- *entry // blocking variant
}

func feedFromLDJSON(file string) (int, error) {
	fd, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	defer fd.Close()

	reader := bufio.NewReader(fd)

	var meta map[string]interface{}
	buf := strings.Builder{}
	cnt := 0
	for {
		substr, isPrefix, err := reader.ReadLine()
		if err == io.EOF {
			break
		}

		buf.WriteString(string(substr))
		if isPrefix {
			continue
		}

		line := buf.String()
		buf.Reset()

		if meta == nil {
			err := json.Unmarshal([]byte(line), &meta)
			if err != nil {
				return 0, err
			}
		} else {
			var entry har.Entry
			err := json.Unmarshal([]byte(line), &entry)
			if err != nil {
				logger.Log.Warningf("Failed decoding entry: %s", line)
			} else {
				cnt += 1
				feedEntry(&entry)
			}
		}
	}

	return cnt, nil
}

func TestFilesList(t *testing.T) {
	res, err := getFiles("./test_artifacts/")
	t.Log(len(res))
	t.Log(res)
	if err != nil || len(res) != 2 {
		t.Logf("Should return 2 files but returned %d", len(res))
		t.FailNow()
	}
}
