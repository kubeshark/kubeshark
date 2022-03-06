package tlstapper

import (
	"bufio"
	"bytes"
	"fmt"
	"net"

	"encoding/binary"
	"encoding/hex"
	"os"
	"strconv"
	"strings"

	"github.com/cilium/ebpf/perf"
	"github.com/go-errors/errors"
	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/tap/api"
)

type tlsPoller struct {
	tls           *TlsTapper
	readers       map[string]*tlsReader
	closedReaders chan string
	reqResMatcher api.RequestResponseMatcher
	chunksReader  *perf.Reader
	extension     *api.Extension
	procfs        string
}

func newTlsPoller(tls *TlsTapper, extension *api.Extension, procfs string) *tlsPoller {
	return &tlsPoller{
		tls:           tls,
		readers:       make(map[string]*tlsReader),
		closedReaders: make(chan string, 100),
		reqResMatcher: extension.Dissector.NewResponseRequestMatcher(),
		extension:     extension,
		chunksReader:  nil,
		procfs:        procfs,
	}
}

func (p *tlsPoller) init(bpfObjects *tlsTapperObjects, bufferSize int) error {
	var err error

	p.chunksReader, err = perf.NewReader(bpfObjects.ChunksBuffer, bufferSize)

	if err != nil {
		return errors.Wrap(err, 0)
	}

	return nil
}

func (p *tlsPoller) close() error {
	return p.chunksReader.Close()
}

func (p *tlsPoller) poll(emitter api.Emitter, options *api.TrafficFilteringOptions) {
	chunks := make(chan *tlsChunk)

	go p.pollChunksPerfBuffer(chunks)

	for {
		select {
		case chunk, ok := <-chunks:
			if !ok {
				return
			}

			if err := p.handleTlsChunk(chunk, p.extension, emitter, options); err != nil {
				LogError(err)
			}
		case key := <-p.closedReaders:
			delete(p.readers, key)
		}
	}
}

func (p *tlsPoller) pollChunksPerfBuffer(chunks chan<- *tlsChunk) {
	logger.Log.Infof("Start polling for tls events")

	for {
		record, err := p.chunksReader.Read()

		if err != nil {
			close(chunks)

			if errors.Is(err, perf.ErrClosed) {
				return
			}

			LogError(errors.Errorf("Error reading chunks from tls perf, aborting TLS! %v", err))
			return
		}

		if record.LostSamples != 0 {
			logger.Log.Infof("Buffer is full, dropped %d chunks", record.LostSamples)
			continue
		}

		buffer := bytes.NewReader(record.RawSample)

		var chunk tlsChunk

		if err := binary.Read(buffer, binary.LittleEndian, &chunk); err != nil {
			LogError(errors.Errorf("Error parsing chunk %v", err))
			continue
		}

		chunks <- &chunk
	}
}

func (p *tlsPoller) handleTlsChunk(chunk *tlsChunk, extension *api.Extension,
	emitter api.Emitter, options *api.TrafficFilteringOptions) error {
	ip, port, err := chunk.getAddress()

	if err != nil {
		return err
	}

	key := buildTlsKey(chunk, ip, port)
	reader, exists := p.readers[key]

	if !exists {
		reader = p.startNewTlsReader(chunk, ip, port, key, extension, emitter, options)
		p.readers[key] = reader
	}

	reader.chunks <- chunk

	if os.Getenv("MIZU_VERBOSE_TLS_TAPPER") == "true" {
		p.logTls(chunk, ip, port)
	}

	return nil
}

func (p *tlsPoller) startNewTlsReader(chunk *tlsChunk, ip net.IP, port uint16, key string, extension *api.Extension,
	emitter api.Emitter, options *api.TrafficFilteringOptions) *tlsReader {

	reader := &tlsReader{
		key:    key,
		chunks: make(chan *tlsChunk, 1),
		doneHandler: func(r *tlsReader) {
			p.closeReader(key, r)
		},
	}

	tcpid := p.buildTcpId(chunk, ip, port)

	go dissect(extension, reader, chunk.isRequest(), &tcpid, emitter, options, p.reqResMatcher)
	return reader
}

func dissect(extension *api.Extension, reader *tlsReader, isRequest bool, tcpid *api.TcpID,
	emitter api.Emitter, options *api.TrafficFilteringOptions, reqResMatcher api.RequestResponseMatcher) {
	b := bufio.NewReader(reader)

	err := extension.Dissector.Dissect(b, isRequest, tcpid, &api.CounterPair{},
		&api.SuperTimer{}, &api.SuperIdentifier{}, emitter, options, reqResMatcher)

	if err != nil {
		logger.Log.Warningf("Error dissecting TLS %v - %v", tcpid, err)
	}
}

func (p *tlsPoller) closeReader(key string, r *tlsReader) {
	close(r.chunks)
	p.closedReaders <- key
}

func buildTlsKey(chunk *tlsChunk, ip net.IP, port uint16) string {
	return fmt.Sprintf("%v:%v-%v:%v", chunk.isClient(), chunk.isRead(), ip, port)
}

func (p *tlsPoller) buildTcpId(chunk *tlsChunk, ip net.IP, port uint16) api.TcpID {
	myIp, myPort, err := getAddressBySockfd(p.procfs, chunk.Pid, chunk.Fd, chunk.isClient())

	if err != nil {
		// May happen if the socket already closed, very likely to happen for localhost
		//
		myIp = api.UnknownIp
		myPort = api.UnknownPort
	}

	if chunk.isRequest() {
		return api.TcpID{
			SrcIP:   myIp.String(),
			DstIP:   ip.String(),
			SrcPort: strconv.FormatUint(uint64(myPort), 10),
			DstPort: strconv.FormatUint(uint64(port), 10),
			Ident:   "",
		}
	} else {
		return api.TcpID{
			SrcIP:   ip.String(),
			DstIP:   myIp.String(),
			SrcPort: strconv.FormatUint(uint64(port), 10),
			DstPort: strconv.FormatUint(uint64(myPort), 10),
			Ident:   "",
		}
	}
}

func (p *tlsPoller) logTls(chunk *tlsChunk, ip net.IP, port uint16) {
	var flagsStr string

	if chunk.isClient() {
		flagsStr = "C"
	} else {
		flagsStr = "S"
	}

	if chunk.isRead() {
		flagsStr += "R"
	} else {
		flagsStr += "W"
	}

	srcIp, srcPort, _ := getAddressBySockfd(p.procfs, chunk.Pid, chunk.Fd, true)
	dstIp, dstPort, _ := getAddressBySockfd(p.procfs, chunk.Pid, chunk.Fd, false)

	str := strings.ReplaceAll(strings.ReplaceAll(string(chunk.Data[0:chunk.Recorded]), "\n", " "), "\r", "")

	logger.Log.Infof("PID: %v (tid: %v) (fd: %v) (client: %v) (addr: %v:%v) (fdaddr %v:%v>%v:%v) (recorded %v out of %v) - %v - %v",
		chunk.Pid, chunk.Tgid, chunk.Fd, flagsStr, ip, port,
		srcIp, srcPort, dstIp, dstPort,
		chunk.Recorded, chunk.Len, str, hex.EncodeToString(chunk.Data[0:chunk.Recorded]))
}
