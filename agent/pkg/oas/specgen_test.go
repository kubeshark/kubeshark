package oas

import (
	"encoding/json"
	"github.com/chanced/openapi"
	"github.com/google/martian/har"
	"github.com/up9inc/mizu/shared/logger"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

// if started via env, write file into subdir
func writeFiles(label string, spec *openapi.OpenAPI) {
	if os.Getenv("MIZU_OAS_WRITE_FILES") != "" {
		path := "./oas-samples"
		err := os.MkdirAll(path, 0o755)
		if err != nil {
			panic(err)
		}
		content, err := json.MarshalIndent(spec, "", "\t")
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile(path+"/"+label+".json", content, 0644)
		if err != nil {
			panic(err)
		}
	}
}

func TestEntries(t *testing.T) {
	files, err := getFiles(".")
	// files, err = getFiles("/media/bigdisk/UP9")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	GetOasGeneratorInstance().Start()

	if err := feedEntries(files); err != nil {
		t.Log(err)
		t.Fail()
	}

	loadStartingOAS()

	svcs := strings.Builder{}
	GetOasGeneratorInstance().ServiceSpecs.Range(func(key, val interface{}) bool {
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

	GetOasGeneratorInstance().ServiceSpecs.Range(func(key, val interface{}) bool {
		svc := key.(string)
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
		writeFiles(svc, spec)

		return true
	})

}

func TestFileLDJSON(t *testing.T) {
	GetOasGeneratorInstance().Start()
	file := "output_rdwtyeoyrj.har.ldjson"
	err := feedFromLDJSON(file)
	if err != nil {
		logger.Log.Warning("Failed processing file: " + err.Error())
		t.Fail()
	}

	loadStartingOAS()

	GetOasGeneratorInstance().ServiceSpecs.Range(func(_, val interface{}) bool {
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

func loadStartingOAS() {
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
	gen.StartFromSpec(doc)

	GetOasGeneratorInstance().ServiceSpecs.Store("catalogue", gen)

	return
}

func TestEntriesNegative(t *testing.T) {
	files := []string{"invalid"}
	err := feedEntries(files)
	if err == nil {
		t.Logf("Should have failed")
		t.Fail()
	}
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
