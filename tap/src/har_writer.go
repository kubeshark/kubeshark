package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/martian/har"
)

const readPermission = 0644
const tempFilenamePrefix = "har_writer"

type PairChanItem struct {
	Request      *http.Request
	RequestTime  time.Time
	Response     *http.Response
	ResponseTime time.Time
}

func openNewHarFile(filename string) *HarFile {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, readPermission)
	if err != nil {
		panic(fmt.Sprintf("Failed to open output file: %s (%v,%+v)", err, err, err))
	}

	harFile := HarFile{file: file, entryCount: 0}
	harFile.writeHeader()

	return &harFile
}

type HarFile struct {
	file *os.File
	entryCount int
}

func NewEntry(request *http.Request, requestTime time.Time, response *http.Response, responseTime time.Time) (*har.Entry, error) {
	// TODO: quick fix until TRA-3212 is implemented  
	if request.URL == nil || request.Method == "" {
		return nil, errors.New("Invalid request")
	}
	harRequest, err := har.NewRequest(request, true)
	if err != nil {
		SilentError("convert-request-to-har", "Failed converting request to HAR %s (%v,%+v)\n", err, err, err)
		return nil, errors.New("Failed converting request to HAR")
	}

	// Martian copies http.Request.URL.String() to har.Request.URL.
	// According to the spec, the URL field needs to be the absolute URL.
	harRequest.URL = fmt.Sprintf("http://%s%s", request.Host, request.URL)

	harResponse, err := har.NewResponse(response, true)
	if err != nil {
		SilentError("convert-response-to-har", "Failed converting response to HAR %s (%v,%+v)\n", err, err, err)
		return nil, errors.New("Failed converting response to HAR")
	}

	totalTime := responseTime.Sub(requestTime).Round(time.Millisecond).Milliseconds()
	if totalTime < 1 {
		totalTime = 1
	}

	harEntry := har.Entry{
		StartedDateTime: time.Now().UTC(),
		Time: totalTime,
		Request: harRequest,
		Response: harResponse,
		Cache: &har.Cache{},
		Timings: &har.Timings{
			Send: -1,
			Wait: -1,
			Receive: totalTime,
		},
	}

	return &harEntry, nil
}

func (f *HarFile) WriteEntry(harEntry *har.Entry) {
	harEntryJson, err := json.Marshal(harEntry)
	if err != nil {
		SilentError("har-entry-marshal", "Failed converting har entry object to JSON%s (%v,%+v)\n", err, err, err)
		return
	}

	var separator string
	if f.GetEntryCount() > 0 {
		separator = ","
	} else {
		separator = ""
	}

	harEntryString := append([]byte(separator), harEntryJson...)

	if _, err := f.file.Write(harEntryString); err != nil {
		panic(fmt.Sprintf("Failed to write to output file: %s (%v,%+v)", err, err, err))
	}

	f.entryCount++
}

func (f *HarFile) GetEntryCount() int {
	return f.entryCount
}

func (f *HarFile) Close() {
	f.writeTrailer()

	err := f.file.Close()
	if err != nil {
		panic(fmt.Sprintf("Failed to close output file: %s (%v,%+v)", err, err, err))
	}
}

func (f*HarFile) writeHeader() {
	header := []byte(`{"log": {"version": "1.2", "creator": {"name": "Mizu", "version": "0.0.1"}, "entries": [`)
	if _, err := f.file.Write(header); err != nil {
		panic(fmt.Sprintf("Failed to write header to output file: %s (%v,%+v)", err, err, err))
	}
}

func (f*HarFile) writeTrailer() {
	trailer := []byte("]}}")
	if _, err := f.file.Write(trailer); err != nil {
		panic(fmt.Sprintf("Failed to write trailer to output file: %s (%v,%+v)", err, err, err))
	}
}

func NewHarWriter(outputDir string, maxEntries int) *HarWriter {
	return &HarWriter{
		OutputDirPath: outputDir,
		MaxEntries: maxEntries,
		PairChan: make(chan *PairChanItem),
		OutChan: make(chan *har.Entry, 1000),
		currentFile: nil,
		done: make(chan bool),
	}
}

type HarWriter struct {
	OutputDirPath string
	MaxEntries int
	PairChan chan *PairChanItem
	OutChan chan *har.Entry
	currentFile *HarFile
	done chan bool
}

func (hw *HarWriter) WritePair(request *http.Request, requestTime time.Time, response *http.Response, responseTime time.Time) {
	hw.PairChan <- &PairChanItem{
		Request: request,
		RequestTime: requestTime,
		Response: response,
		ResponseTime: responseTime,
	}
}

func (hw *HarWriter) Start() {
	if hw.OutputDirPath != "" {
		if err := os.MkdirAll(hw.OutputDirPath, os.ModePerm); err != nil {
			panic(fmt.Sprintf("Failed to create output directory: %s (%v,%+v)", err, err, err))
		}
	}

	go func() {
		for pair := range hw.PairChan {
			harEntry, err := NewEntry(pair.Request, pair.RequestTime, pair.Response, pair.ResponseTime)
			if err != nil {
				continue
			}

			if hw.OutputDirPath != "" {
				if hw.currentFile == nil {
					hw.openNewFile()
				}

				hw.currentFile.WriteEntry(harEntry)

				if hw.currentFile.GetEntryCount() >= hw.MaxEntries {
					hw.closeFile()
				}
			} else {
				hw.OutChan <- harEntry
			}
		}

		if hw.currentFile != nil {
			hw.closeFile()
		}
		hw.done <- true
	} ()
}

func (hw *HarWriter) Stop() {
	close(hw.PairChan)
	<-hw.done
}

func (hw *HarWriter) openNewFile() {
	filename := filepath.Join(os.TempDir(), fmt.Sprintf("%s_%d", tempFilenamePrefix, time.Now().UnixNano()))
	hw.currentFile = openNewHarFile(filename)
}

func (hw *HarWriter) closeFile() {
	hw.currentFile.Close()
	tmpFilename := hw.currentFile.file.Name()
	hw.currentFile = nil

	filename := buildFilename(hw.OutputDirPath, time.Now())
	os.Rename(tmpFilename, filename)
}

func buildFilename(dir string, t time.Time) string {
	// (epoch time in nanoseconds)__(YYYY_Month_DD__hh-mm-ss).har
	filename := fmt.Sprintf("%d__%s.har", t.UnixNano(), t.Format("2006_Jan_02__15-04-05"))
	return filepath.Join(dir, filename)
}
