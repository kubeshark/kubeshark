package oas

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/chanced/openapi"
	"github.com/up9inc/mizu/agent/pkg/har"
	"github.com/up9inc/mizu/logger"
	"github.com/wI2L/jsondiff"
)

// if started via env, write file into subdir
func outputSpec(label string, spec *openapi.OpenAPI, t *testing.T) string {
	content, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		panic(err)
	}

	if os.Getenv("MIZU_OAS_WRITE_FILES") != "" {
		path := "./oas-samples"
		err := os.MkdirAll(path, 0o755)
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile(path+"/"+label+".json", content, 0644)
		if err != nil {
			panic(err)
		}
		t.Logf("Written: %s", label)
	} else {
		t.Logf("%s", string(content))
	}
	return string(content)
}

func TestEntries(t *testing.T) {
	//logger.InitLoggerStd(logging.INFO) causes race condition
	files, err := getFiles("./test_artifacts/")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	gen := NewDefaultOasGenerator()
	gen.serviceSpecs = new(sync.Map)
	loadStartingOAS("test_artifacts/catalogue.json", "catalogue", gen.serviceSpecs)
	loadStartingOAS("test_artifacts/trcc.json", "trcc-api-service", gen.serviceSpecs)

	go func() {
		for {
			time.Sleep(1 * time.Second)
			gen.serviceSpecs.Range(func(key, val interface{}) bool {
				svc := key.(string)
				t.Logf("Getting spec for %s", svc)
				gen := val.(*SpecGen)
				_, err := gen.GetSpec()
				if err != nil {
					t.Error(err)
				}
				return true
			})
		}
	}()

	cnt, err := feedEntries(files, true, gen)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	svcs := strings.Builder{}
	gen.serviceSpecs.Range(func(key, val interface{}) bool {
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

	gen.serviceSpecs.Range(func(key, val interface{}) bool {
		svc := key.(string)
		gen := val.(*SpecGen)
		spec, err := gen.GetSpec()
		if err != nil {
			t.Log(err)
			t.FailNow()
		}

		outputSpec(svc, spec, t)

		err = spec.Validate()
		if err != nil {
			t.Log(err)
			t.FailNow()
		}

		return true
	})

	logger.Log.Infof("Total entries: %d", cnt)
}

func TestFileSingle(t *testing.T) {
	gen := NewDefaultOasGenerator()
	gen.serviceSpecs = new(sync.Map)
	// loadStartingOAS()
	file := "test_artifacts/params.har"
	files := []string{file}
	cnt, err := feedEntries(files, true, gen)
	if err != nil {
		logger.Log.Warning("Failed processing file: " + err.Error())
		t.Fail()
	}

	gen.serviceSpecs.Range(func(key, val interface{}) bool {
		svc := key.(string)
		gen := val.(*SpecGen)
		spec, err := gen.GetSpec()
		if err != nil {
			t.Log(err)
			t.FailNow()
		}

		specText := outputSpec(svc, spec, t)

		err = spec.Validate()
		if err != nil {
			t.Log(err)
			t.FailNow()
		}

		expected, err := ioutil.ReadFile(file + ".spec.json")
		if err != nil {
			t.Errorf(err.Error())
			t.FailNow()
		}

		patFloatPrecision := regexp.MustCompile(`(\d+\.\d{1,2})(\d*)`)

		expected = []byte(patUuid4.ReplaceAllString(string(expected), "<UUID4>"))
		specText = patUuid4.ReplaceAllString(specText, "<UUID4>")
		expected = []byte(patFloatPrecision.ReplaceAllString(string(expected), "$1"))
		specText = patFloatPrecision.ReplaceAllString(specText, "$1")

		diff, err := jsondiff.CompareJSON(expected, []byte(specText))
		if err != nil {
			t.Errorf(err.Error())
			t.FailNow()
		}

		if os.Getenv("MIZU_OAS_WRITE_FILES") != "" {
			err = ioutil.WriteFile(file+".spec.json", []byte(specText), 0644)
			if err != nil {
				panic(err)
			}
		}

		if len(diff) > 0 {
			t.Errorf("Generated spec does not match expected:\n%s", diff.String())
		}

		return true
	})

	logger.Log.Infof("Processed entries: %d", cnt)
}

func loadStartingOAS(file string, label string, specs *sync.Map) {
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

	gen := NewGen(label)
	gen.StartFromSpec(doc)

	specs.Store(label, gen)
}

func TestEntriesNegative(t *testing.T) {
	gen := NewDefaultOasGenerator()
	gen.serviceSpecs = new(sync.Map)
	files := []string{"invalid"}
	_, err := feedEntries(files, false, gen)
	if err == nil {
		t.Logf("Should have failed")
		t.Fail()
	}
}

func TestEntriesPositive(t *testing.T) {
	gen := NewDefaultOasGenerator()
	gen.serviceSpecs = new(sync.Map)
	files := []string{"test_artifacts/params.har"}
	_, err := feedEntries(files, false, gen)
	if err != nil {
		t.Logf("Failed")
		t.Fail()
	}
}

func TestLoadValidHAR(t *testing.T) {
	inp := `{"startedDateTime": "2021-02-03T07:48:12.959000+00:00", "time": 1, "request": {"method": "GET", "url": "http://unresolved_target/1.0.0/health", "httpVersion": "HTTP/1.1", "cookies": [], "headers": [], "queryString": [], "headersSize": -1, "bodySize": -1}, "response": {"status": 200, "statusText": "OK", "httpVersion": "HTTP/1.1", "cookies": [], "headers": [], "content": {"size": 2, "mimeType": "", "text": "OK"}, "redirectURL": "", "headersSize": -1, "bodySize": 2}, "cache": {}, "timings": {"send": -1, "wait": -1, "receive": 1}}`
	var entry *har.Entry
	var err = json.Unmarshal([]byte(inp), &entry)
	if err != nil {
		t.Logf("Failed to decode entry: %s", err)
		t.FailNow() // demonstrates the problem of `martian` HAR library
	}
}

func TestLoadValid3_1(t *testing.T) {
	fd, err := os.Open("test_artifacts/catalogue.json")
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
}
