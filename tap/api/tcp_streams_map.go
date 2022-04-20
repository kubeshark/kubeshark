package api

import (
	"runtime"
	"sync"
	"time"

	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/tap/api/diagnose"
)

type TcpStreamMap struct {
	Streams  *sync.Map
	streamId int64
}

func NewTcpStreamMap() *TcpStreamMap {
	return &TcpStreamMap{
		Streams: &sync.Map{},
	}
}

func (streamMap *TcpStreamMap) Store(key, value interface{}) {
	streamMap.Streams.Store(key, value)
}

func (streamMap *TcpStreamMap) Delete(key interface{}) {
	streamMap.Streams.Delete(key)
}

func (streamMap *TcpStreamMap) NextId() int64 {
	streamMap.streamId++
	return streamMap.streamId
}

func (streamMap *TcpStreamMap) CloseTimedoutTcpStreamChannels() {
	tcpStreamChannelTimeoutMs := GetTcpChannelTimeoutMs()
	closeTimedoutTcpChannelsIntervalMs := GetCloseTimedoutTcpChannelsInterval()
	logger.Log.Infof("Using %d ms as the close timedout TCP stream channels interval", closeTimedoutTcpChannelsIntervalMs/time.Millisecond)

	ticker := time.NewTicker(closeTimedoutTcpChannelsIntervalMs)
	for {
		<-ticker.C

		streamMap.Streams.Range(func(key interface{}, value interface{}) bool {
			stream := value.(*TcpStream)
			if stream.ProtoIdentifier.Protocol == nil {
				if !stream.isClosed && time.Now().After(stream.createdAt.Add(tcpStreamChannelTimeoutMs)) {
					stream.Close()
					diagnose.AppStatsInst.IncDroppedTcpStreams()
					logger.Log.Debugf("Dropped an unidentified TCP stream because of timeout. Total dropped: %d Total Goroutines: %d Timeout (ms): %d",
						diagnose.AppStatsInst.DroppedTcpStreams, runtime.NumGoroutine(), tcpStreamChannelTimeoutMs/time.Millisecond)
				}
			}
			return true
		})
	}
}
