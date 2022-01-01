package oas

import (
	"encoding/json"
	"github.com/chanced/openapi"
	har "github.com/mrichman/hargo"
	"github.com/op/go-logging"
	"github.com/up9inc/mizu/shared/logger"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func init() {
	logger.InitLoggerStderrOnly(logging.DEBUG)
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

	specs := sync.Map{}
	finished := false
	mutex := sync.Mutex{}
	go func() { // this goroutine generates OAS from entries
		err := EntriesToSpecs(entries, &specs)

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
			spec, err := gen.getSpec()
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

		t.Logf("Made a cycle on %d specs: %s", svcs.Len(), svcs.String())

		mutex.Lock()
		if finished {
			mutex.Unlock()
			break
		}
		mutex.Unlock()
	}

	specs.Range(func(_, val interface{}) bool {
		gen := val.(*SpecGen)
		spec, err := gen.getSpec()
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
