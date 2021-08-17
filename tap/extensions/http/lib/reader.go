package lib

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/bradleyfalzon/tlsx"
)

const checkTLSPacketAmount = 100

type httpReaderDataMsg struct {
	bytes     []byte
	timestamp time.Time
}

type tcpID struct {
	srcIP   string
	dstIP   string
	srcPort string
	dstPort string
}

type ConnectionInfo struct {
	ClientIP   string
	ClientPort string
	ServerIP   string
	ServerPort string
	IsOutgoing bool
}

func SomeExample() {
	fmt.Println("Hello from SomeExample")
}

func (tid *tcpID) String() string {
	return fmt.Sprintf("%s->%s %s->%s", tid.srcIP, tid.dstIP, tid.srcPort, tid.dstPort)
}

/* httpReader gets reads from a channel of bytes of tcp payload, and parses it into HTTP/1 requests and responses.
 * The payload is written to the channel by a tap.TcpStream object that is dedicated to one tcp connection.
 * An httpReader object is unidirectional: it parses either a client stream or a server stream.
 * Implements io.Reader interface (Read)
 */
type httpReader struct {
	ident              string
	tcpID              tcpID
	isClient           bool
	isHTTP2            bool
	isOutgoing         bool
	msgQueue           chan httpReaderDataMsg // Channel of captured reassembled tcp payload
	data               []byte
	captureTime        time.Time
	hexdump            bool
	grpcAssembler      GrpcAssembler
	messageCount       uint
	harWriter          *HarWriter
	packetsSeen        uint
	outboundLinkWriter *OutboundLinkWriter
}

func (h *httpReader) Read(p []byte) (int, error) {
	var msg httpReaderDataMsg

	ok := true
	for ok && len(h.data) == 0 {
		msg, ok = <-h.msgQueue
		h.data = msg.bytes

		h.captureTime = msg.timestamp
		if len(h.data) > 0 {
			h.packetsSeen += 1
		}
		if h.packetsSeen < checkTLSPacketAmount && len(msg.bytes) > 5 { // packets with less than 5 bytes cause tlsx to panic
			clientHello := tlsx.ClientHello{}
			err := clientHello.Unmarshall(msg.bytes)
			if err == nil {
				fmt.Printf("Detected TLS client hello with SNI %s\n", clientHello.SNI)
				numericPort, _ := strconv.Atoi(h.tcpID.dstPort)
				h.outboundLinkWriter.WriteOutboundLink(h.tcpID.srcIP, h.tcpID.dstIP, numericPort, clientHello.SNI, TLSProtocol)
			}
		}
	}
	if !ok || len(h.data) == 0 {
		return 0, io.EOF
	}

	l := copy(p, h.data)
	h.data = h.data[l:]
	return l, nil
}

func (h *httpReader) run(wg *sync.WaitGroup) {
	defer wg.Done()
	b := bufio.NewReader(h)

	if isHTTP2, err := checkIsHTTP2Connection(b, h.isClient); err != nil {
		SilentError("HTTP/2-Prepare-Connection", "stream %s Failed to check if client is HTTP/2: %s (%v,%+v)", h.ident, err, err, err)
		// Do something?
	} else {
		h.isHTTP2 = isHTTP2
	}

	if h.isHTTP2 {
		err := prepareHTTP2Connection(b, h.isClient)
		if err != nil {
			SilentError("HTTP/2-Prepare-Connection-After-Check", "stream %s error: %s (%v,%+v)", h.ident, err, err, err)
		}
		h.grpcAssembler = createGrpcAssembler(b)
	}

	for true {
		if h.isHTTP2 {
			err := h.handleHTTP2Stream()
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else if err != nil {
				SilentError("HTTP/2", "stream %s error: %s (%v,%+v)", h.ident, err, err, err)
				continue
			}
		} else if h.isClient {
			err := h.handleHTTP1ClientStream(b)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else if err != nil {
				SilentError("HTTP-request", "stream %s Request error: %s (%v,%+v)", h.ident, err, err, err)
				continue
			}
		} else {
			err := h.handleHTTP1ServerStream(b)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else if err != nil {
				SilentError("HTTP-response", "stream %s Response error: %s (%v,%+v)", h.ident, err, err, err)
				continue
			}
		}
	}
}

func (h *httpReader) handleHTTP2Stream() error {
	streamID, messageHTTP1, err := h.grpcAssembler.readMessage()
	h.messageCount++
	if err != nil {
		return err
	}

	var reqResPair *requestResponsePair
	var connectionInfo *ConnectionInfo

	switch messageHTTP1 := messageHTTP1.(type) {
	case http.Request:
		ident := fmt.Sprintf("%s->%s %s->%s %d", h.tcpID.srcIP, h.tcpID.dstIP, h.tcpID.srcPort, h.tcpID.dstPort, streamID)
		connectionInfo = &ConnectionInfo{
			ClientIP:   h.tcpID.srcIP,
			ClientPort: h.tcpID.srcPort,
			ServerIP:   h.tcpID.dstIP,
			ServerPort: h.tcpID.dstPort,
			IsOutgoing: h.isOutgoing,
		}
		reqResPair = reqResMatcher.registerRequest(ident, &messageHTTP1, h.captureTime)
	case http.Response:
		ident := fmt.Sprintf("%s->%s %s->%s %d", h.tcpID.dstIP, h.tcpID.srcIP, h.tcpID.dstPort, h.tcpID.srcPort, streamID)
		connectionInfo = &ConnectionInfo{
			ClientIP:   h.tcpID.dstIP,
			ClientPort: h.tcpID.dstPort,
			ServerIP:   h.tcpID.srcIP,
			ServerPort: h.tcpID.srcPort,
			IsOutgoing: h.isOutgoing,
		}
		reqResPair = reqResMatcher.registerResponse(ident, &messageHTTP1, h.captureTime)
	}

	if reqResPair != nil {
		// statsTracker.incMatchedMessages()

		if h.harWriter != nil {
			h.harWriter.WritePair(
				reqResPair.Request.orig.(*http.Request),
				reqResPair.Request.captureTime,
				reqResPair.Response.orig.(*http.Response),
				reqResPair.Response.captureTime,
				connectionInfo,
			)
		}
	}

	return nil
}

func (h *httpReader) handleHTTP1ClientStream(b *bufio.Reader) error {
	req, err := http.ReadRequest(b)
	h.messageCount++
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(req.Body)
	req.Body = io.NopCloser(bytes.NewBuffer(body)) // rewind
	s := len(body)
	if err != nil {
		SilentError("HTTP-request-body", "stream %s Got body err: %s", h.ident, err)
	} else if h.hexdump {
		Debug("Body(%d/0x%x) - %s", len(body), len(body), hex.Dump(body))
	}
	if err := req.Body.Close(); err != nil {
		SilentError("HTTP-request-body-close", "stream %s Failed to close request body: %s", h.ident, err)
	}
	encoding := req.Header["Content-Encoding"]
	Debug("HTTP/1 Request: %s %s %s (Body:%d) -> %s", h.ident, req.Method, req.URL, s, encoding)

	ident := fmt.Sprintf("%s->%s %s->%s %d", h.tcpID.srcIP, h.tcpID.dstIP, h.tcpID.srcPort, h.tcpID.dstPort, h.messageCount)
	reqResPair := reqResMatcher.registerRequest(ident, req, h.captureTime)
	if reqResPair != nil {
		// statsTracker.incMatchedMessages()

		if h.harWriter != nil {
			h.harWriter.WritePair(
				reqResPair.Request.orig.(*http.Request),
				reqResPair.Request.captureTime,
				reqResPair.Response.orig.(*http.Response),
				reqResPair.Response.captureTime,
				&ConnectionInfo{
					ClientIP:   h.tcpID.srcIP,
					ClientPort: h.tcpID.srcPort,
					ServerIP:   h.tcpID.dstIP,
					ServerPort: h.tcpID.dstPort,
					IsOutgoing: h.isOutgoing,
				},
			)
		}
	}

	// h.parent.Lock()
	// h.parent.urls = append(h.parent.urls, req.URL.String())
	// h.parent.Unlock()

	return nil
}

func (h *httpReader) handleHTTP1ServerStream(b *bufio.Reader) error {
	res, err := http.ReadResponse(b, nil)
	h.messageCount++
	var req string
	// h.parent.Lock()
	// if len(h.parent.urls) == 0 {
	// 	req = fmt.Sprintf("<no-request-seen>")
	// } else {
	// 	req, h.parent.urls = h.parent.urls[0], h.parent.urls[1:]
	// }
	// h.parent.Unlock()
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body = io.NopCloser(bytes.NewBuffer(body)) // rewind
	s := len(body)
	if err != nil {
		SilentError("HTTP-response-body", "HTTP/%s: failed to get body(parsed len:%d): %s", h.ident, s, err)
	}
	if h.hexdump {
		Debug("Body(%d/0x%x) - %s", len(body), len(body), hex.Dump(body))
	}
	if err := res.Body.Close(); err != nil {
		SilentError("HTTP-response-body-close", "HTTP/%s: failed to close body(parsed len:%d): %s", h.ident, s, err)
	}
	sym := ","
	if res.ContentLength > 0 && res.ContentLength != int64(s) {
		sym = "!="
	}
	contentType, ok := res.Header["Content-Type"]
	if !ok {
		contentType = []string{http.DetectContentType(body)}
	}
	encoding := res.Header["Content-Encoding"]
	Debug("HTTP/1 Response: %s %s URL:%s (%d%s%d%s) -> %s", h.ident, res.Status, req, res.ContentLength, sym, s, contentType, encoding)

	ident := fmt.Sprintf("%s->%s %s->%s %d", h.tcpID.dstIP, h.tcpID.srcIP, h.tcpID.dstPort, h.tcpID.srcPort, h.messageCount)
	reqResPair := reqResMatcher.registerResponse(ident, res, h.captureTime)
	if reqResPair != nil {
		// statsTracker.incMatchedMessages()

		if h.harWriter != nil {
			h.harWriter.WritePair(
				reqResPair.Request.orig.(*http.Request),
				reqResPair.Request.captureTime,
				reqResPair.Response.orig.(*http.Response),
				reqResPair.Response.captureTime,
				&ConnectionInfo{
					ClientIP:   h.tcpID.dstIP,
					ClientPort: h.tcpID.dstPort,
					ServerIP:   h.tcpID.srcIP,
					ServerPort: h.tcpID.srcPort,
					IsOutgoing: h.isOutgoing,
				},
			)
		}
	}

	return nil
}
