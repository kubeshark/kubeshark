package tap

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/martian/har"
)

const readPermission = 0644
const harFilenameSuffix = ".har"
const tempFilenameSuffix = ".har.tmp"

type PairChanItem struct {
	Request         *http.Request
	RequestTime     time.Time
	Response        *http.Response
	ResponseTime    time.Time
	RequestSenderIp string
	ConnectionInfo  *ConnectionInfo
}

func openNewHarFile(filename string) *HarFile {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, readPermission)
	if err != nil {
		log.Panicf("Failed to open output file: %s (%v,%+v)", err, err, err)
	}

	harFile := HarFile{file: file, entryCount: 0}
	harFile.writeHeader()

	return &harFile
}

type HarFile struct {
	file       *os.File
	entryCount int
}

func NewEntry(request *http.Request, requestTime time.Time, response *http.Response, responseTime time.Time) (*har.Entry, error) {
	harRequest, err := har.NewRequest(request, false)
	if err != nil {
		SilentError("convert-request-to-har", "Failed converting request to HAR %s (%v,%+v)", err, err, err)
		return nil, errors.New("Failed converting request to HAR")
	}

	// For requests with multipart/form-data or application/x-www-form-urlencoded Content-Type,
	// martian/har will parse the request body and place the parameters in harRequest.PostData.Params
	// instead of harRequest.PostData.Text (as the HAR spec requires it).
	// Mizu currently only looks at PostData.Text. Therefore, instead of letting martian/har set the content of
	// PostData, always copy the request body to PostData.Text.
	if (request.ContentLength > 0) {
		reqBody, err := ioutil.ReadAll(request.Body)
		if err != nil {
			SilentError("read-request-body", "Failed converting request to HAR %s (%v,%+v)", err, err, err)
			return nil, errors.New("Failed reading request body")
		}
		request.Body = ioutil.NopCloser(bytes.NewReader(reqBody))
		harRequest.PostData.Text = string(reqBody)
	}

	harResponse, err := har.NewResponse(response, true)
	if err != nil {
		SilentError("convert-response-to-har", "Failed converting response to HAR %s (%v,%+v)", err, err, err)
		return nil, errors.New("Failed converting response to HAR")
	}

	if harRequest.PostData != nil && strings.HasPrefix(harRequest.PostData.MimeType, "application/grpc") {
		// Force HTTP/2 gRPC into HAR template

		harRequest.URL = fmt.Sprintf("%s://%s%s", request.Header.Get(":scheme"), request.Header.Get(":authority"), request.Header.Get(":path"))

		status, err := strconv.Atoi(response.Header.Get(":status"))
		if err != nil {
			SilentError("convert-response-status-for-har", "Failed converting status to int %s (%v,%+v)", err, err, err)
			return nil, errors.New("Failed converting response status to int for HAR")
		}
		harResponse.Status = status
	} else {
		// Martian copies http.Request.URL.String() to har.Request.URL, which usually contains the path.
		// However, according to the HAR spec, the URL field needs to be the absolute URL.
		var scheme string
		if request.URL.Scheme != "" {
			scheme = request.URL.Scheme
		} else {
			scheme = "http"
		}
		harRequest.URL = fmt.Sprintf("%s://%s%s", scheme, request.Host, request.URL)
	}

	totalTime := responseTime.Sub(requestTime).Round(time.Millisecond).Milliseconds()
	if totalTime < 1 {
		totalTime = 1
	}

	harEntry := har.Entry{
		StartedDateTime: time.Now().UTC(),
		Time:            totalTime,
		Request:         harRequest,
		Response:        harResponse,
		Cache:           &har.Cache{},
		Timings: &har.Timings{
			Send:    -1,
			Wait:    -1,
			Receive: totalTime,
		},
	}

	return &harEntry, nil
}

func (f *HarFile) WriteEntry(harEntry *har.Entry) {
	harEntryJson, err := json.Marshal(harEntry)
	if err != nil {
		SilentError("har-entry-marshal", "Failed converting har entry object to JSON%s (%v,%+v)", err, err, err)
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
		log.Panicf("Failed to write to output file: %s (%v,%+v)", err, err, err)
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
		log.Panicf("Failed to close output file: %s (%v,%+v)", err, err, err)
	}
}

func (f *HarFile) writeHeader() {
	header := []byte(`{"log": {"version": "1.2", "creator": {"name": "Mizu", "version": "0.0.1"}, "entries": [`)
	if _, err := f.file.Write(header); err != nil {
		log.Panicf("Failed to write header to output file: %s (%v,%+v)", err, err, err)
	}
}

func (f *HarFile) writeTrailer() {
	trailer := []byte("]}}")
	if _, err := f.file.Write(trailer); err != nil {
		log.Panicf("Failed to write trailer to output file: %s (%v,%+v)", err, err, err)
	}
}

func NewHarWriter(outputDir string, maxEntries int) *HarWriter {
	return &HarWriter{
		OutputDirPath: outputDir,
		MaxEntries:    maxEntries,
		PairChan:      make(chan *PairChanItem),
		OutChan:       make(chan *OutputChannelItem, 1000),
		currentFile:   nil,
		done:          make(chan bool),
	}
}

type OutputChannelItem struct {
	HarEntry               *har.Entry
	ConnectionInfo         *ConnectionInfo
	ValidationRulesChecker string
}

type HarWriter struct {
	OutputDirPath string
	MaxEntries    int
	PairChan      chan *PairChanItem
	OutChan       chan *OutputChannelItem
	currentFile   *HarFile
	done          chan bool
}

func (hw *HarWriter) WritePair(request *http.Request, requestTime time.Time, response *http.Response, responseTime time.Time, connectionInfo *ConnectionInfo) {
	hw.PairChan <- &PairChanItem{
		Request:        request,
		RequestTime:    requestTime,
		Response:       response,
		ResponseTime:   responseTime,
		ConnectionInfo: connectionInfo,
	}
}

func (hw *HarWriter) Start() {
	if hw.OutputDirPath != "" {
		if err := os.MkdirAll(hw.OutputDirPath, os.ModePerm); err != nil {
			log.Panicf("Failed to create output directory: %s (%v,%+v)", err, err, err)
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
				hw.OutChan <- &OutputChannelItem{
					HarEntry:       harEntry,
					ConnectionInfo: pair.ConnectionInfo,
				}
			}
		}

		if hw.currentFile != nil {
			hw.closeFile()
		}
		hw.done <- true
	}()
}

func (hw *HarWriter) Stop() {
	close(hw.PairChan)
	<-hw.done
	close(hw.OutChan)
}

func (hw *HarWriter) openNewFile() {
	filename := buildFilename(hw.OutputDirPath, time.Now(), tempFilenameSuffix)
	hw.currentFile = openNewHarFile(filename)
}

func (hw *HarWriter) closeFile() {
	hw.currentFile.Close()
	tmpFilename := hw.currentFile.file.Name()
	hw.currentFile = nil

	filename := buildFilename(hw.OutputDirPath, time.Now(), harFilenameSuffix)
	err := os.Rename(tmpFilename, filename)
	if err != nil {
		SilentError("Rename-file", "cannot rename file: %s (%v,%+v)", err, err, err)
	}
}

func buildFilename(dir string, t time.Time, suffix string) string {
	// (epoch time in nanoseconds)__(YYYY_Month_DD__hh-mm-ss).har
	filename := fmt.Sprintf("%d__%s%s", t.UnixNano(), t.Format("2006_Jan_02__15-04-05"), suffix)
	return filepath.Join(dir, filename)
}
