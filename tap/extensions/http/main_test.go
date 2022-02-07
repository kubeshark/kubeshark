package http

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/up9inc/mizu/tap/api"
)

const (
	pattern    = "./bin/*_req.bin"
	msg        = "Dissecting:"
	respSuffix = "_res.bin"
	expectDir  = "expect"
	testUpdate = "TEST_UPDATE"
)

func TestDissect(t *testing.T) {
	var testUpdateEnabled bool
	_, present := os.LookupEnv(testUpdate)
	if present {
		testUpdateEnabled = true
	}

	if testUpdateEnabled {
		os.RemoveAll(expectDir)
		err := os.MkdirAll(expectDir, 0775)
		assert.Nil(t, err)
	}

	dissector := NewDissector()
	paths, err := filepath.Glob(pattern)
	if err != nil {
		log.Fatal(err)
	}

	options := &api.TrafficFilteringOptions{
		IgnoredUserAgents: []string{},
	}

	for _, path := range paths {
		basePath := path[:len(path)-8]

		// Channel to verify the output
		itemChannel := make(chan *api.OutputChannelItem)
		var emitter api.Emitter = &api.Emitting{
			AppStats:      &api.AppStats{},
			OutputChannel: itemChannel,
		}

		var items []*api.OutputChannelItem

		go func() {
			for item := range itemChannel {
				items = append(items, item)
			}
		}()

		// Stream level
		counterPair := &api.CounterPair{
			Request:  0,
			Response: 0,
		}
		superIdentifier := &api.SuperIdentifier{}

		// Request
		pathClient := path
		fmt.Printf("%s %s\n", msg, pathClient)
		fileClient, err := os.Open(pathClient)
		assert.Nil(t, err)

		bufferClient := bufio.NewReader(fileClient)
		tcpIDClient := &api.TcpID{
			SrcIP:   "1",
			DstIP:   "2",
			SrcPort: "1",
			DstPort: "2",
		}
		err = dissector.Dissect(bufferClient, true, tcpIDClient, counterPair, &api.SuperTimer{}, superIdentifier, emitter, options)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			panic(err)
		}

		// Response
		pathServer := basePath + respSuffix
		fmt.Printf("%s %s\n", msg, pathServer)
		fileServer, err := os.Open(pathServer)
		assert.Nil(t, err)

		bufferServer := bufio.NewReader(fileServer)
		tcpIDServer := &api.TcpID{
			SrcIP:   "2",
			DstIP:   "1",
			SrcPort: "2",
			DstPort: "1",
		}
		err = dissector.Dissect(bufferServer, false, tcpIDServer, counterPair, &api.SuperTimer{}, superIdentifier, emitter, options)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			panic(err)
		}

		fileClient.Close()
		fileServer.Close()

		time.Sleep(10 * time.Millisecond)

		pathExpect := fmt.Sprintf("%s/%s.json", expectDir, basePath[4:])

		sort.Slice(items, func(i, j int) bool {
			iMarshaled, err := json.Marshal(items[i])
			assert.Nil(t, err)
			jMarshaled, err := json.Marshal(items[j])
			assert.Nil(t, err)
			return len(iMarshaled) < len(jMarshaled)
		})

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
