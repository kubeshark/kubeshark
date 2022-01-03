package oas

import (
	"encoding/json"
	"github.com/chanced/openapi"
	"github.com/google/martian/har"
	"github.com/up9inc/mizu/shared/logger"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

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
