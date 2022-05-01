package oas

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/up9inc/mizu/agent/pkg/har"

	"github.com/up9inc/mizu/logger"
)

func getFiles(baseDir string) (result []string, err error) {
	result = make([]string, 0)
	logger.Log.Infof("Reading files from tree: %s", baseDir)

	inputs := []string{baseDir}

	// https://yourbasic.org/golang/list-files-in-directory/
	visitor := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Mode()&os.ModeSymlink != 0 {
			path, _ = os.Readlink(path)
			inputs = append(inputs, path)
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !info.IsDir() && (ext == ".har" || ext == ".ldjson") {
			result = append(result, path)
		}

		return nil
	}

	for len(inputs) > 0 {
		path := inputs[0]
		inputs = inputs[1:]
		err = filepath.Walk(path, visitor)
	}

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

func feedEntries(fromFiles []string, isSync bool, gen *defaultOasGenerator) (count uint, err error) {
	badFiles := make([]string, 0)
	cnt := uint(0)
	for _, file := range fromFiles {
		logger.Log.Info("Processing file: " + file)
		ext := strings.ToLower(filepath.Ext(file))
		eCnt := uint(0)
		switch ext {
		case ".har":
			eCnt, err = feedFromHAR(file, isSync, gen)
			if err != nil {
				logger.Log.Warning("Failed processing file: " + err.Error())
				badFiles = append(badFiles, file)
				continue
			}
		case ".ldjson":
			eCnt, err = feedFromLDJSON(file, isSync, gen)
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

func feedFromHAR(file string, isSync bool, gen *defaultOasGenerator) (uint, error) {
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

	cnt := uint(0)
	for _, entry := range harDoc.Log.Entries {
		cnt += 1
		feedEntry(&entry, "", file, gen, fmt.Sprintf("%024d", cnt))
	}

	return cnt, nil
}

func feedEntry(entry *har.Entry, source string, file string, gen *defaultOasGenerator, cnt string) {
	entry.Comment = file
	if entry.Response.Status == 302 {
		logger.Log.Debugf("Dropped traffic entry due to permanent redirect status: %s", entry.StartedDateTime)
	}

	if strings.Contains(entry.Request.URL, "some") { // for debugging
		logger.Log.Debugf("Interesting: %s", entry.Request.URL)
	}

	u, err := url.Parse(entry.Request.URL)
	if err != nil {
		logger.Log.Errorf("Failed to parse entry URL: %v, err: %v", entry.Request.URL, err)
	}

	ews := EntryWithSource{Entry: *entry, Source: source, Destination: u.Host, Id: cnt}
	gen.handleHARWithSource(&ews)
}

func feedFromLDJSON(file string, isSync bool, gen *defaultOasGenerator) (uint, error) {
	fd, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	defer fd.Close()

	reader := bufio.NewReader(fd)

	var meta map[string]interface{}
	buf := strings.Builder{}
	cnt := uint(0)
	source := ""
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
			if s, ok := meta["_source"]; ok && s != nil {
				source = s.(string)
			}
		} else {
			var entry har.Entry
			err := json.Unmarshal([]byte(line), &entry)
			if err != nil {
				logger.Log.Warningf("Failed decoding entry: %s", line)
			} else {
				cnt += 1
				feedEntry(&entry, source, file, gen, fmt.Sprintf("%024d", cnt))
			}
		}
	}

	return cnt, nil
}

func TestFilesList(t *testing.T) {
	res, err := getFiles("./test_artifacts/")
	t.Log(len(res))
	t.Log(res)
	if err != nil || len(res) != 3 {
		t.Logf("Should return 2 files but returned %d", len(res))
		t.FailNow()
	}
}
