package tlstapper

import (
	"bufio"
	"fmt"
	"net"

	"encoding/hex"
	"os"
	"strconv"
	"strings"

	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/tap/api"
)

const UNKNOWN_PORT uint16 = 80
const UNKNOWN_HOST string = "127.0.0.1"

type tlsPoller struct {
	tls           *TlsTapper
	readers       map[string]*tlsReader
	closedReaders chan string
}

func NewTlsPoller(tls *TlsTapper) *tlsPoller {
	return &tlsPoller{
		tls:           tls,
		readers:       make(map[string]*tlsReader),
		closedReaders: make(chan string, 100),
	}
}

func (p *tlsPoller) Poll(httpExtension *api.Extension,
	emitter api.Emitter, options *api.TrafficFilteringOptions) {

	chunks := make(chan *tlsChunk)

	go p.tls.pollPerf(chunks)

	for {
		select {
		case chunk := <-chunks:
			p.handleTlsChunk(chunk, httpExtension, emitter, options)
		case key := <-p.closedReaders:
			delete(p.readers, key)
		}
	}
}

func (p *tlsPoller) handleTlsChunk(chunk *tlsChunk, httpExtension *api.Extension,
	emitter api.Emitter, options *api.TrafficFilteringOptions) error {
	ip, port, err := chunk.getAddress()

	if err != nil {
		logger.Log.Warningf("Error getting address from tls chunk %v", err)
		return err
	}

	key := buildTlsKey(chunk, ip, port)
	reader, exists := p.readers[key]

	if !exists {
		reader = p.startNewTlsReader(chunk, ip, port, key, httpExtension, emitter, options)
		p.readers[key] = reader
	}

	reader.chunks <- chunk

	if os.Getenv("MIZU_VERBOSE_TLS_TAPPER") == "true" {
		logTls(chunk, ip, port)
	}

	return nil
}

func (p *tlsPoller) startNewTlsReader(chunk *tlsChunk, ip net.IP, port uint16, key string, httpExtension *api.Extension,
	emitter api.Emitter, options *api.TrafficFilteringOptions) *tlsReader {

	reader := &tlsReader{
		key:    key,
		chunks: make(chan *tlsChunk, 1),
		doneHandler: func(r *tlsReader) {
			p.closeReader(key, r)
		},
	}

	isRequest := (chunk.isClient() && chunk.isWrite()) || (chunk.isServer() && chunk.isRead())
	tcpid := buildTcpId(isRequest, ip, port)
	b := bufio.NewReader(reader)
	go httpExtension.Dissector.Dissect(b, isRequest, &tcpid, &api.CounterPair{}, &api.SuperTimer{}, &api.SuperIdentifier{}, emitter, options)
	return reader
}

func (p *tlsPoller) closeReader(key string, r *tlsReader) {
	close(r.chunks)
	p.closedReaders <- key
}

func buildTlsKey(chunk *tlsChunk, ip net.IP, port uint16) string {
	return fmt.Sprintf("%v:%v-%v:%v", chunk.isClient(), chunk.isRead(), ip, port)
}

func buildTcpId(isRequest bool, ip net.IP, port uint16) api.TcpID {
	if isRequest {
		return api.TcpID{
			SrcIP:   UNKNOWN_HOST,
			DstIP:   ip.String(),
			SrcPort: strconv.Itoa(int(UNKNOWN_PORT)),
			DstPort: strconv.FormatInt(int64(port), 10),
			Ident:   "",
		}
	} else {
		return api.TcpID{
			SrcIP:   ip.String(),
			DstIP:   UNKNOWN_HOST,
			SrcPort: strconv.FormatInt(int64(port), 10),
			DstPort: strconv.Itoa(int(UNKNOWN_PORT)),
			Ident:   "",
		}
	}
}

func logTls(chunk *tlsChunk, ip net.IP, port uint16) {
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

	logger.Log.Infof("PID: %v (tid: %v) (fd: %v) (client: %v) (addr: %v:%v) (recorded %v out of %v) - %v - %v",
		chunk.Pid, chunk.Tgid, chunk.Fd, flagsStr, ip, port, chunk.Recorded, chunk.Len, str, hex.EncodeToString(chunk.Data[0:chunk.Recorded]))
}
