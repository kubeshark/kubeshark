package oas

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/mrichman/hargo"
	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/tap/api"
	"io"
	"io/ioutil"
	"mizuserver/pkg/utils"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func getFiles(baseDir string) (result []string, err error) {
	result = make([]string, 0, 0)

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

	return result, err
}

func feedEntries(fromFiles []string, out chan<- hargo.Entry) (err error) {
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

func feedFromHAR(file string, out chan<- hargo.Entry) error {
	fd, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	defer fd.Close()

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return err
	}

	var har hargo.Har
	err = json.Unmarshal(data, &har)
	if err != nil {
		return err
	}

	for _, entry := range har.Log.Entries {
		out <- entry
	}

	return nil
}

func feedFromLDJSON(file string, out chan<- hargo.Entry) error {
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
			var entry hargo.Entry
			err := json.Unmarshal([]byte(line), &entry)
			if err != nil {
				return err
			}
			out <- entry
		}
	}

	return nil
}

func EntriesToSpecs(entries <-chan *api.MizuEntry, specs *sync.Map) error {
	for {
		mizuEntry, ok := <-entries
		if !ok {
			break
		}

		if mizuEntry.Protocol.Name != "http" {
			logger.Log.Debugf("Skipped non-HTTP entry for now: %s/%s", mizuEntry.Protocol.Name, mizuEntry.Id)
			continue // TODO: handle non-HTTP entries into AsyncAPI specs
		}

		entry, err := utils.NewEntry(mizuEntry.Request, mizuEntry.Response, mizuEntry.StartTime, mizuEntry.ElapsedTime)
		if err != nil {
			return err
		}

		u, err := url.Parse(entry.Request.URL)
		if err != nil {
			return err
		}

		val, found := specs.Load(u.Host)
		var gen SpecGen
		if !found {
			gen = *NewGen(u.Host)
			specs.Store(u.Host, gen)
		} else {
			gen = val.(SpecGen)
		}

		err = gen.feedEntry(entry)
		if err != nil {
			return err
		}
	}
	return nil
}
