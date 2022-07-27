package redis

import (
	"fmt"
	"time"

	"github.com/up9inc/mizu/tap/api"
)

func handleClientStream(progress *api.ReadProgress, capture api.Capture, tcpID *api.TcpID, counterPair *api.CounterPair, captureTime time.Time, emitter api.Emitter, request *RedisPacket, reqResMatcher *requestResponseMatcher) error {
	counterPair.Lock()
	counterPair.Request++
	requestCounter := counterPair.Request
	counterPair.Unlock()

	ident := fmt.Sprintf(
		"%s_%s_%s_%s_%d",
		tcpID.SrcIP,
		tcpID.DstIP,
		tcpID.SrcPort,
		tcpID.DstPort,
		requestCounter,
	)

	item := reqResMatcher.registerRequest(ident, request, captureTime, progress.Current())
	if item != nil {
		item.Capture = capture
		item.ConnectionInfo = &api.ConnectionInfo{
			ClientIP:   tcpID.SrcIP,
			ClientPort: tcpID.SrcPort,
			ServerIP:   tcpID.DstIP,
			ServerPort: tcpID.DstPort,
			IsOutgoing: true,
		}
		emitter.Emit(item)
	}
	return nil
}

func handleServerStream(progress *api.ReadProgress, capture api.Capture, tcpID *api.TcpID, counterPair *api.CounterPair, captureTime time.Time, emitter api.Emitter, response *RedisPacket, reqResMatcher *requestResponseMatcher) error {
	counterPair.Lock()
	counterPair.Response++
	responseCounter := counterPair.Response
	counterPair.Unlock()

	ident := fmt.Sprintf(
		"%s_%s_%s_%s_%d",
		tcpID.DstIP,
		tcpID.SrcIP,
		tcpID.DstPort,
		tcpID.SrcPort,
		responseCounter,
	)

	item := reqResMatcher.registerResponse(ident, response, captureTime, progress.Current())
	if item != nil {
		item.Capture = capture
		item.ConnectionInfo = &api.ConnectionInfo{
			ClientIP:   tcpID.DstIP,
			ClientPort: tcpID.DstPort,
			ServerIP:   tcpID.SrcIP,
			ServerPort: tcpID.SrcPort,
			IsOutgoing: false,
		}
		emitter.Emit(item)
	}
	return nil
}
