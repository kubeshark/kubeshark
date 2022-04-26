package tlstapper

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"sync"
	"time"

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
	tls            *TlsTapper
	readers        map[string]*tlsReader
	closedReaders  chan string
	reqResMatcher  api.RequestResponseMatcher
	chunksReader   *perf.Reader
	extension      *api.Extension
	procfs         string
	pidToNamespace sync.Map
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

	p.updateTlsReader(chunk, reader, ip, port)
	
	if os.Getenv("MIZU_VERBOSE_TLS_TAPPER") == "true" {
		p.logTls(chunk, key, reader)
	}

	return nil
}

func (p *tlsPoller) updateTlsReader(chunk *tlsChunk, reader *tlsReader, ip net.IP, port uint16) {
	myIp, myPort, err := getAddressBySockfd(p.procfs, chunk.Pid, chunk.Fd, chunk.isClient())

	if err == nil {
		// Currently tls key is built from the foreign address, pid, thread id and more, it doesn't contains 
		//	the local address though. The reason is pure technical, and require traversing kernel structs in
		//	the ebpf (hopefully we'll get to fix it).
		//
		// If there are many tcp requests to the same remote address, at around the same time, from different
		//	connections, they all get the same tls key.
		//
		// If we don't update tcpID with the new local address, the requests matcher fails, because it have
		//	the old tcp id used by the previous connection.
		//
		// Yet getAddressBySockfd may fail because the file descriptor already gone, usually for localhost 
		// 	where everything is CPU, or for the last chunk in a big message. If that happen, its better to 
		//	stay with the old tcpID which most chances is still valid, instead of replacing it with unknown.
		//
		reader.tcpID = p.buildTcpId(chunk, ip, port, myIp, myPort)
	}
	
	reader.newChunk(chunk)
}

func (p *tlsPoller) startNewTlsReader(chunk *tlsChunk, ip net.IP, port uint16, key string, extension *api.Extension,
	emitter api.Emitter, options *api.TrafficFilteringOptions) *tlsReader {

	reader := &tlsReader{
		key:    key,
		chunks: make(chan *tlsChunk, 1),
		doneHandler: func(r *tlsReader) {
			p.closeReader(key, r)
		},
		progress: &api.ReadProgress{},
		timer: api.SuperTimer{
			CaptureTime: time.Now(),
		},
		tcpID: p.buildInitialTcpID(chunk, ip, port),
	}

	tlsEmitter := &tlsEmitter{
		delegate:  emitter,
		namespace: p.getNamespace(chunk.Pid),
	}

	go dissect(extension, reader, chunk.isRequest(), tlsEmitter, options, p.reqResMatcher)
	return reader
}

func dissect(extension *api.Extension, reader *tlsReader, isRequest bool,
	tlsEmitter *tlsEmitter, options *api.TrafficFilteringOptions, reqResMatcher api.RequestResponseMatcher) {
	b := bufio.NewReader(reader)

	err := extension.Dissector.Dissect(b, reader.progress, api.Ebpf, isRequest, &reader.tcpID, &api.CounterPair{},
		&reader.timer, &api.SuperIdentifier{}, tlsEmitter, options, reqResMatcher)

	if err != nil {
		logger.Log.Warningf("Error dissecting TLS %v - %v", reader.tcpID, err)
	}
}

func (p *tlsPoller) closeReader(key string, r *tlsReader) {
	close(r.chunks)
	p.closedReaders <- key
}

func buildTlsKey(chunk *tlsChunk, ip net.IP, port uint16) string {
	var clientStr string
	if chunk.isClient() {
		clientStr = "C"
	} else {
		clientStr = "S"
	}

	var readerStr string
	if chunk.isRead() {
		readerStr = "R"
	} else {
		readerStr = "W"
	}
	
	// This same ID used by the ebpf C code
	//
	id := uint64(chunk.Pid) | uint64(chunk.Tgid) << 32

	return fmt.Sprintf("%d-%s%s-%v:%d", id, clientStr, readerStr, ip, port)
}

func (p *tlsPoller) buildInitialTcpID(chunk *tlsChunk, ip net.IP, port uint16) api.TcpID {
	myIp, myPort, err := getAddressBySockfd(p.procfs, chunk.Pid, chunk.Fd, chunk.isClient())

	if err != nil {
		// May happen if the socket already closed, very likely to happen for localhost during testing,
		//  we prefer to fallback to unkonwn address than dropping.
		//
		myIp = api.UnknownIp
		myPort = api.UnknownPort
	}

	return p.buildTcpId(chunk, ip, port, myIp, myPort)
}

func (p *tlsPoller) buildTcpId(chunk *tlsChunk, ip net.IP, port uint16, myIp net.IP, myPort uint16) api.TcpID {
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

func (p *tlsPoller) addPid(pid uint32, namespace string) {
	p.pidToNamespace.Store(pid, namespace)
}

func (p *tlsPoller) getNamespace(pid uint32) string {
	namespaceIfc, ok := p.pidToNamespace.Load(pid)

	if !ok {
		return api.UNKNOWN_NAMESPACE
	}

	namespace, ok := namespaceIfc.(string)

	if !ok {
		return api.UNKNOWN_NAMESPACE
	}

	return namespace
}

func (p *tlsPoller) clearPids() {
	p.pidToNamespace.Range(func(key, v interface{}) bool {
		p.pidToNamespace.Delete(key)
		return true
	})
}

func (p *tlsPoller) logTls(chunk *tlsChunk, key string, reader *tlsReader) {
	srcIp, srcPort, _ := getAddressBySockfd(p.procfs, chunk.Pid, chunk.Fd, true)
	dstIp, dstPort, _ := getAddressBySockfd(p.procfs, chunk.Pid, chunk.Fd, false)

	str := strings.ReplaceAll(strings.ReplaceAll(string(chunk.Data[0:chunk.Recorded]), "\n", " "), "\r", "")

	logger.Log.Infof("[%-32s] #%-4d (fd: %d %s:%d>%s:%d) (recorded %d/%d:%d) (tcpid: %s) - %s - %s",
		key, reader.seenChunks, chunk.Fd, srcIp, srcPort, dstIp, dstPort,
		chunk.Recorded, chunk.Len, chunk.Start, reader.tcpID,
		str, hex.EncodeToString(chunk.Data[0:chunk.Recorded]))
}
