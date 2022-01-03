package oas

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/chanced/openapi"
	"github.com/google/martian/har"
	"github.com/up9inc/mizu/shared/logger"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
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

func feedEntries(fromFiles []string, out chan har.Entry) (err error) {
	defer close(out)

	for _, file := range fromFiles {
		logger.Log.Info("Processing file: " + file)
		ext := strings.ToLower(filepath.Ext(file))
		switch ext {
		case ".har":
			err = feedFromHAR(file, out)
			if err != nil {
				logger.Log.Warning("Failed processing file: " + err.Error())
				continue
			}
		case ".ldjson":
			err = feedFromLDJSON(file, out)
			if err != nil {
				logger.Log.Warning("Failed processing file: " + err.Error())
				continue
			}
		default:
			return errors.New("Unsupported file extension: " + ext)
		}
	}

	return nil
}

func feedFromHAR(file string, out chan<- har.Entry) error {
	fd, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	defer fd.Close()

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return err
	}

	var harDoc har.HAR
	err = json.Unmarshal(data, &harDoc)
	if err != nil {
		return err
	}

	for _, entry := range harDoc.Log.Entries {
		out <- *entry
	}

	return nil
}

func feedFromLDJSON(file string, out chan<- har.Entry) error {
	fd, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	defer fd.Close()

	reader := bufio.NewReader(fd)

	var meta map[string]interface{}

	buf := strings.Builder{}
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
				return err
			}
		} else {
			var entry har.Entry
			err := json.Unmarshal([]byte(line), &entry)
			if err != nil {
				logger.Log.Warningf("Failed decoding entry: %s", line)
			}
			out <- entry
		}
	}

	return nil
}

func TestFilesList(t *testing.T) {
	res, err := getFiles(".")
	t.Log(len(res))
	t.Log(res)
	if err != nil || len(res) != 2 {
		t.Logf("Should return 2 files but returned %d", len(res))
		t.FailNow()
	}
}

func TestEntries(t *testing.T) {
	files, err := getFiles(".")
	// files, err = getFiles("/media/bigdisk/UP9")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	entries := make(chan har.Entry)
	go func() { // this goroutine reads inputs
		err := feedEntries(files, entries)
		if err != nil {
			t.Log(err)
			t.Fail()
		}
	}()

	specs := new(sync.Map)

	loadStartingOAS(specs)

	finished := false
	mutex := sync.Mutex{}
	go func() { // this goroutine generates OAS from entries
		err := EntriesToSpecs(entries, specs)

		mutex.Lock()
		finished = true
		mutex.Unlock()

		if err != nil {
			t.Log(err)
			t.Fail()
		}
	}()

	for { // demo for parallel fetching of specs
		time.Sleep(time.Second / 2)
		svcs := strings.Builder{}
		specs.Range(func(key, val interface{}) bool {
			gen := val.(*SpecGen)
			svc := key.(string)
			svcs.WriteString(svc + ",")
			spec, err := gen.GetSpec()
			if err != nil {
				t.Log(err)
				t.FailNow()
				return false
			}

			err = spec.Validate()
			if err != nil {
				specText, _ := json.MarshalIndent(spec, "", "\t")
				t.Log(string(specText))
				t.Log(err)
				t.FailNow()
			}

			return true
		})

		t.Logf("Made a cycle on specs: %s", svcs.String())

		mutex.Lock()
		if finished {
			mutex.Unlock()
			break
		}
		mutex.Unlock()
	}

	specs.Range(func(_, val interface{}) bool {
		gen := val.(*SpecGen)
		spec, err := gen.GetSpec()
		if err != nil {
			t.Log(err)
			t.FailNow()
		}

		specText, _ := json.MarshalIndent(spec, "", "\t")
		t.Logf("%s", string(specText))

		err = spec.Validate()
		if err != nil {
			t.Log(err)
			t.FailNow()
		}

		return true
	})

}

func TestFileLDJSON(t *testing.T) {
	entries := make(chan har.Entry)
	go func() {
		file := "output_rdwtyeoyrj.har.ldjson"
		err := feedFromLDJSON(file, entries)
		if err != nil {
			logger.Log.Warning("Failed processing file: " + err.Error())
			t.Fail()
		}
		close(entries)
	}()

	specs := new(sync.Map)

	loadStartingOAS(specs)

	err := EntriesToSpecs(entries, specs)
	if err != nil {
		logger.Log.Warning("Failed processing entries: " + err.Error())
		t.FailNow()
	}

	specs.Range(func(_, val interface{}) bool {
		gen := val.(*SpecGen)
		spec, err := gen.GetSpec()
		if err != nil {
			t.Log(err)
			t.FailNow()
		}

		specText, _ := json.MarshalIndent(spec, "", "\t")
		t.Logf("%s", string(specText))

		err = spec.Validate()
		if err != nil {
			t.Log(err)
			t.FailNow()
		}

		return true
	})
}

func loadStartingOAS(specs *sync.Map) {
	file := "catalogue.json"
	fd, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	defer fd.Close()

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		panic(err)
	}

	var doc *openapi.OpenAPI
	err = json.Unmarshal(data, &doc)
	if err != nil {
		panic(err)
	}

	gen := NewGen("catalogue")
	gen.startFromSpec(doc)

	specs.Store("catalogue", gen)

	return
}

func TestEntriesNegative(t *testing.T) {
	files := []string{"invalid"}
	entries := make(chan har.Entry)
	go func() {
		err := feedEntries(files, entries)
		if err == nil {
			t.Logf("Should have failed")
			t.Fail()
		}
	}()
}

func TestLoadValidHAR(t *testing.T) {
	inp := `{"startedDateTime": "2021-02-03T07:48:12.959000+00:00", "time": 1, "request": {"method": "GET", "url": "http://unresolved_target/1.0.0/health", "httpVersion": "HTTP/1.1", "cookies": [], "headers": [], "queryString": [], "headersSize": -1, "bodySize": -1}, "response": {"status": 200, "statusText": "OK", "httpVersion": "HTTP/1.1", "cookies": [], "headers": [], "content": {"size": 2, "mimeType": "", "text": "OK"}, "redirectURL": "", "headersSize": -1, "bodySize": 2}, "cache": {}, "timings": {"send": -1, "wait": -1, "receive": 1}}`
	var entry *har.Entry
	var err = json.Unmarshal([]byte(inp), &entry)
	if err != nil {
		t.Logf("Failed to decode entry: %s", err)
		// t.FailNow() demonstrates the problem of library
	}
}

func TestLoadValid3_1(t *testing.T) {
	fd, err := os.Open("catalogue.json")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	defer fd.Close()

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	var oas openapi.OpenAPI
	err = json.Unmarshal(data, &oas)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	return
}
