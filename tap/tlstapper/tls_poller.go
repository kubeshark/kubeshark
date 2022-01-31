package tlstapper

import (
	"bufio"
	"bytes"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/tap/api"
)

const UNKWONW_PORT string = "80"
const UNKWONW_HOST string = "127.0.0.1"

func Poll(tls *TlsTapper, httpExtension *api.Extension,
	emitter api.Emitter, options *api.TrafficFilteringOptions) {

	chunks := make(chan *tlsChunk)

	go tls.pollPerf(chunks)

	for {
		chunk := <-chunks

		ip, port, err := chunk.getAddress()

		if err != nil {
			logger.Log.Warningf("Error getting address from tls chunk %v", err)
			continue
		}

		var id api.TcpID

		reader := bufio.NewReader(bytes.NewReader(chunk.Data[0:chunk.Recorded]))
		isRequest := (chunk.isClient() && chunk.isWrite()) || (chunk.isServer() && chunk.isRead())
		
		if isRequest {
			id = api.TcpID{
				SrcIP:   UNKWONW_HOST,
				DstIP:   ip.String(),
				SrcPort: UNKWONW_PORT,
				DstPort: strconv.FormatInt(int64(port), 10),
			}
		} else {
			id = api.TcpID{
				SrcIP:   ip.String(),
				DstIP:   UNKNOWN_HOST,
				SrcPort: strconv.FormatInt(int64(port), 10),
				DstPort: UNKNOWN_PORT,
			}
		}

		httpExtension.Dissector.Dissect(reader, isRequest, &id, &api.CounterPair{}, &api.SuperTimer{}, &api.SuperIdentifier{}, emitter, options)

		if os.Getenv("MIZU_VERBOSE_TLS_TAPPER") == "true" {
			logTls(chunk, ip, port)
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

	logger.Log.Infof("PID: %v (tid: %v) (fd: %v) (client: %v) (addr: %v:%v) (recorded %v out of %v) - %v",
		chunk.Pid, chunk.Tgid, chunk.Fd, flagsStr, ip, port, chunk.Recorded, chunk.Len, str)
}
