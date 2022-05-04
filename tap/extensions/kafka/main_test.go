package kafka

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/up9inc/mizu/tap/api"
)

const (
	binDir          = "bin"
	patternBin      = "*_req.bin"
	patternExpect   = "*.json"
	msgDissecting   = "Dissecting:"
	msgAnalyzing    = "Analyzing:"
	msgSummarizing  = "Summarizing:"
	msgRepresenting = "Representing:"
	respSuffix      = "_res.bin"
	expectDir       = "expect"
	dissectDir      = "dissect"
	analyzeDir      = "analyze"
	summarizeDir    = "summarize"
	representDir    = "represent"
	testUpdate      = "TEST_UPDATE"
)

func TestRegister(t *testing.T) {
	dissector := NewDissector()
	extension := &api.Extension{}
	dissector.Register(extension)
	assert.Equal(t, "kafka", extension.Protocol.Name)
}

func TestMacros(t *testing.T) {
	expectedMacros := map[string]string{
		"kafka": `proto.name == "kafka"`,
	}
	dissector := NewDissector()
	macros := dissector.Macros()
	assert.Equal(t, expectedMacros, macros)
}

func TestPing(t *testing.T) {
	dissector := NewDissector()
	dissector.Ping()
}

func TestDissect(t *testing.T) {
	_, testUpdateEnabled := os.LookupEnv(testUpdate)

	expectDirDissect := path.Join(expectDir, dissectDir)

	if testUpdateEnabled {
		os.RemoveAll(expectDirDissect)
		err := os.MkdirAll(expectDirDissect, 0775)
		assert.Nil(t, err)
	}

	dissector := NewDissector()
	paths, err := filepath.Glob(path.Join(binDir, patternBin))
	if err != nil {
		log.Fatal(err)
	}

	options := &api.TrafficFilteringOptions{
		IgnoredUserAgents: []string{},
	}

	for _, _path := range paths {
		basePath := _path[:len(_path)-8]

		// Channel to verify the output
		itemChannel := make(chan *api.OutputChannelItem)
		var emitter api.Emitter = &api.Emitting{
			AppStats:      &api.AppStats{},
			OutputChannel: itemChannel,
		}

		var items []*api.OutputChannelItem
		stop := make(chan bool)

		go func() {
			for {
				select {
				case <-stop:
					return
				case item := <-itemChannel:
					items = append(items, item)
				}
			}
		}()

		// Stream level
		counterPair := &api.CounterPair{
			Request:  0,
			Response: 0,
		}

		// Request
		pathClient := _path
		fmt.Printf("%s %s\n", msgDissecting, pathClient)
		fileClient, err := os.Open(pathClient)
		assert.Nil(t, err)

		bufferClient := bufio.NewReader(fileClient)
		tcpIDClient := &api.TcpID{
			SrcIP:   "1",
			DstIP:   "2",
			SrcPort: "1",
			DstPort: "2",
		}
		reqResMatcher := dissector.NewResponseRequestMatcher()
		reqResMatcher.SetMaxTry(10)
		stream := NewTcpStream(api.Pcap)
		reader := NewTcpReader(
			&api.ReadProgress{},
			"",
			tcpIDClient,
			time.Time{},
			stream,
			true,
			false,
			nil,
			emitter,
			counterPair,
			reqResMatcher,
		)
		err = dissector.Dissect(bufferClient, reader, options)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println(err)
		}

		// Response
		pathServer := basePath + respSuffix
		fmt.Printf("%s %s\n", msgDissecting, pathServer)
		fileServer, err := os.Open(pathServer)
		assert.Nil(t, err)

		bufferServer := bufio.NewReader(fileServer)
		tcpIDServer := &api.TcpID{
			SrcIP:   "2",
			DstIP:   "1",
			SrcPort: "2",
			DstPort: "1",
		}
		reader = NewTcpReader(
			&api.ReadProgress{},
			"",
			tcpIDServer,
			time.Time{},
			stream,
			false,
			false,
			nil,
			emitter,
			counterPair,
			reqResMatcher,
		)
		err = dissector.Dissect(bufferServer, reader, options)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println(err)
		}

		fileClient.Close()
		fileServer.Close()

		pathExpect := path.Join(expectDirDissect, fmt.Sprintf("%s.json", basePath[4:]))

		time.Sleep(10 * time.Millisecond)

		stop <- true

		marshaled, err := json.Marshal(items)
		assert.Nil(t, err)

		if testUpdateEnabled {
			if len(items) > 0 {
				err = os.WriteFile(pathExpect, marshaled, 0644)
				assert.Nil(t, err)
			}
		} else {
			if _, err := os.Stat(pathExpect); errors.Is(err, os.ErrNotExist) {
				assert.Len(t, items, 0)
			} else {
				expectedBytes, err := ioutil.ReadFile(pathExpect)
				assert.Nil(t, err)

				assert.JSONEq(t, string(expectedBytes), string(marshaled))
			}
		}
	}
}

func TestAnalyze(t *testing.T) {
	_, testUpdateEnabled := os.LookupEnv(testUpdate)

	expectDirDissect := path.Join(expectDir, dissectDir)
	expectDirAnalyze := path.Join(expectDir, analyzeDir)

	if testUpdateEnabled {
		os.RemoveAll(expectDirAnalyze)
		err := os.MkdirAll(expectDirAnalyze, 0775)
		assert.Nil(t, err)
	}

	dissector := NewDissector()
	paths, err := filepath.Glob(path.Join(expectDirDissect, patternExpect))
	if err != nil {
		log.Fatal(err)
	}

	for _, _path := range paths {
		fmt.Printf("%s %s\n", msgAnalyzing, _path)

		bytes, err := ioutil.ReadFile(_path)
		assert.Nil(t, err)

		var items []*api.OutputChannelItem
		err = json.Unmarshal(bytes, &items)
		assert.Nil(t, err)

		var entries []*api.Entry
		for _, item := range items {
			entry := dissector.Analyze(item, "", "", "")
			entries = append(entries, entry)
		}

		pathExpect := path.Join(expectDirAnalyze, filepath.Base(_path))

		marshaled, err := json.Marshal(entries)
		assert.Nil(t, err)

		if testUpdateEnabled {
			if len(entries) > 0 {
				err = os.WriteFile(pathExpect, marshaled, 0644)
				assert.Nil(t, err)
			}
		} else {
			if _, err := os.Stat(pathExpect); errors.Is(err, os.ErrNotExist) {
				assert.Len(t, items, 0)
			} else {
				expectedBytes, err := ioutil.ReadFile(pathExpect)
				assert.Nil(t, err)

				assert.JSONEq(t, string(expectedBytes), string(marshaled))
			}
		}
	}
}

func TestSummarize(t *testing.T) {
	_, testUpdateEnabled := os.LookupEnv(testUpdate)

	expectDirAnalyze := path.Join(expectDir, analyzeDir)
	expectDirSummarize := path.Join(expectDir, summarizeDir)

	if testUpdateEnabled {
		os.RemoveAll(expectDirSummarize)
		err := os.MkdirAll(expectDirSummarize, 0775)
		assert.Nil(t, err)
	}

	dissector := NewDissector()
	paths, err := filepath.Glob(path.Join(expectDirAnalyze, patternExpect))
	if err != nil {
		log.Fatal(err)
	}

	for _, _path := range paths {
		fmt.Printf("%s %s\n", msgSummarizing, _path)

		bytes, err := ioutil.ReadFile(_path)
		assert.Nil(t, err)

		var entries []*api.Entry
		err = json.Unmarshal(bytes, &entries)
		assert.Nil(t, err)

		var baseEntries []*api.BaseEntry
		for _, entry := range entries {
			baseEntry := dissector.Summarize(entry)
			baseEntries = append(baseEntries, baseEntry)
		}

		pathExpect := path.Join(expectDirSummarize, filepath.Base(_path))

		marshaled, err := json.Marshal(baseEntries)
		assert.Nil(t, err)

		if testUpdateEnabled {
			if len(baseEntries) > 0 {
				err = os.WriteFile(pathExpect, marshaled, 0644)
				assert.Nil(t, err)
			}
		} else {
			if _, err := os.Stat(pathExpect); errors.Is(err, os.ErrNotExist) {
				assert.Len(t, entries, 0)
			} else {
				expectedBytes, err := ioutil.ReadFile(pathExpect)
				assert.Nil(t, err)

				assert.JSONEq(t, string(expectedBytes), string(marshaled))
			}
		}
	}
}

func TestRepresent(t *testing.T) {
	_, testUpdateEnabled := os.LookupEnv(testUpdate)

	expectDirAnalyze := path.Join(expectDir, analyzeDir)
	expectDirRepresent := path.Join(expectDir, representDir)

	if testUpdateEnabled {
		os.RemoveAll(expectDirRepresent)
		err := os.MkdirAll(expectDirRepresent, 0775)
		assert.Nil(t, err)
	}

	dissector := NewDissector()
	paths, err := filepath.Glob(path.Join(expectDirAnalyze, patternExpect))
	if err != nil {
		log.Fatal(err)
	}

	for _, _path := range paths {
		fmt.Printf("%s %s\n", msgRepresenting, _path)

		bytes, err := ioutil.ReadFile(_path)
		assert.Nil(t, err)

		var entries []*api.Entry
		err = json.Unmarshal(bytes, &entries)
		assert.Nil(t, err)

		var objects []string
		for _, entry := range entries {
			object, err := dissector.Represent(entry.Request, entry.Response)
			assert.Nil(t, err)
			objects = append(objects, string(object))
		}

		pathExpect := path.Join(expectDirRepresent, filepath.Base(_path))

		marshaled, err := json.Marshal(objects)
		assert.Nil(t, err)

		if testUpdateEnabled {
			if len(objects) > 0 {
				err = os.WriteFile(pathExpect, marshaled, 0644)
				assert.Nil(t, err)
			}
		} else {
			if _, err := os.Stat(pathExpect); errors.Is(err, os.ErrNotExist) {
				assert.Len(t, objects, 0)
			} else {
				expectedBytes, err := ioutil.ReadFile(pathExpect)
				assert.Nil(t, err)

				assert.JSONEq(t, string(expectedBytes), string(marshaled))
			}
		}
	}
}
