package tlstapper

import (
	"bufio"
	"bytes"
	"fmt"
	"sync"
	"time"

	"encoding/binary"
	"encoding/hex"
	"os"
	"strconv"
	"strings"

	"github.com/cilium/ebpf/perf"
	"github.com/go-errors/errors"
	"github.com/up9inc/mizu/logger"
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
	fdCache        map[string]addressPair
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
		fdCache:       make(map[string]addressPair),
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
	address, err := p.getAddressPair(chunk)

	if err != nil {
		address, err = p.getFallbackAddress(chunk)

		if err != nil {
			return err
		}
	}

	key := buildTlsKey(address)
	reader, exists := p.readers[key]

	if !exists {
		reader = p.startNewTlsReader(chunk, &address, key, emitter, extension, options)
		p.readers[key] = reader
	}

	reader.newChunk(chunk)

	if os.Getenv("MIZU_VERBOSE_TLS_TAPPER") == "true" {
		p.logTls(chunk, key, reader)
	}

	return nil
}

func (p *tlsPoller) startNewTlsReader(chunk *tlsChunk, address *addressPair, key string,
	emitter api.Emitter, extension *api.Extension, options *api.TrafficFilteringOptions) *tlsReader {

	tcpid := p.buildTcpId(chunk, address)

	doneHandler := func(r *tlsReader) {
		p.closeReader(key, r)
	}

	tlsEmitter := &tlsEmitter{
		delegate:  emitter,
		namespace: p.getNamespace(chunk.Pid),
	}

	reader := &tlsReader{
		key:           key,
		chunks:        make(chan *tlsChunk, 1),
		doneHandler:   doneHandler,
		progress:      &api.ReadProgress{},
		tcpID:         &tcpid,
		isClient:      chunk.isRequest(),
		captureTime:   time.Now(),
		extension:     extension,
		emitter:       tlsEmitter,
		counterPair:   &api.CounterPair{},
		reqResMatcher: p.reqResMatcher,
	}

	stream := &tlsStream{
		reader:          reader,
		protoIdentifier: &api.ProtoIdentifier{},
	}

	reader.parent = stream

	go dissect(extension, reader, options)
	return reader
}

func dissect(extension *api.Extension, reader *tlsReader, options *api.TrafficFilteringOptions) {
	b := bufio.NewReader(reader)

	err := extension.Dissector.Dissect(b, reader, options)

	if err != nil {
		logger.Log.Warningf("Error dissecting TLS %v - %v", reader.GetTcpID(), err)
	}
}

func (p *tlsPoller) closeReader(key string, r *tlsReader) {
	close(r.chunks)
	p.closedReaders <- key
}

func (p *tlsPoller) getAddressPair(chunk *tlsChunk) (addressPair, error) {
	address, err := getAddressBySockfd(p.procfs, chunk.Pid, chunk.Fd)
	fdCacheKey := fmt.Sprintf("%d:%d", chunk.Pid, chunk.Fd)

	if err == nil {
		if !chunk.isRequest() {
			switchedAddress := addressPair{
				srcIp:   address.dstIp,
				srcPort: address.dstPort,
				dstIp:   address.srcIp,
				dstPort: address.srcPort,
			}
			p.fdCache[fdCacheKey] = switchedAddress
			return switchedAddress, nil
		} else {
			p.fdCache[fdCacheKey] = address
			return address, nil
		}
	}

	fromCache, ok := p.fdCache[fdCacheKey]

	if !ok {
		return addressPair{}, err
	}

	return fromCache, nil
}

func (p *tlsPoller) getFallbackAddress(chunk *tlsChunk) (addressPair, error) {
	ip, port, err := chunk.getAddress()

	if err != nil {
		return addressPair{}, err
	}

	if chunk.isRequest() {
		return addressPair{
			srcIp:   api.UnknownIp,
			srcPort: api.UnknownPort,
			dstIp:   ip,
			dstPort: port,
		}, nil
	} else {
		return addressPair{
			srcIp:   ip,
			srcPort: port,
			dstIp:   api.UnknownIp,
			dstPort: api.UnknownPort,
		}, nil
	}
}

func buildTlsKey(address addressPair) string {
	return fmt.Sprintf("%s:%d>%s:%d", address.srcIp, address.srcPort, address.dstIp, address.dstPort)
}

func (p *tlsPoller) buildTcpId(chunk *tlsChunk, address *addressPair) api.TcpID {
	return api.TcpID{
		SrcIP:   address.srcIp.String(),
		DstIP:   address.dstIp.String(),
		SrcPort: strconv.FormatUint(uint64(address.srcPort), 10),
		DstPort: strconv.FormatUint(uint64(address.dstPort), 10),
		Ident:   "",
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

	str := strings.ReplaceAll(strings.ReplaceAll(string(chunk.Data[0:chunk.Recorded]), "\n", " "), "\r", "")

	logger.Log.Infof("[%-44s] %s #%-4d (fd: %d) (recorded %d/%d:%d) - %s - %s",
		key, flagsStr, reader.seenChunks, chunk.Fd,
		chunk.Recorded, chunk.Len, chunk.Start,
		str, hex.EncodeToString(chunk.Data[0:chunk.Recorded]))
}
