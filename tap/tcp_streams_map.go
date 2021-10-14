package tap

import (
	"runtime"
	_debug "runtime/debug"
	"sync"
	"time"

	"github.com/romana/rlog"
)

type tcpStreamMap struct {
	streams  *sync.Map
	streamId int64
}

func NewTcpStreamMap() *tcpStreamMap {
	return &tcpStreamMap{
		streams: &sync.Map{},
	}
}

func (streamMap *tcpStreamMap) Store(key, value interface{}) {
	streamMap.streams.Store(key, value)
}

func (streamMap *tcpStreamMap) Delete(key interface{}) {
	streamMap.streams.Delete(key)
}

func (streamMap *tcpStreamMap) nextId() int64 {
	streamMap.streamId++
	return streamMap.streamId
}

func (streamMap *tcpStreamMap) closeTimedoutTcpStreamChannels() {
	tcpStreamChannelTimeout := GetTcpChannelTimeoutMs()
	for {
		time.Sleep(10 * time.Millisecond)
		_debug.FreeOSMemory()
		streamMap.streams.Range(func(key interface{}, value interface{}) bool {
			streamWrapper := value.(*tcpStreamWrapper)
			stream := streamWrapper.stream
			if stream.superIdentifier.Protocol == nil {
				if !stream.isClosed && time.Now().After(streamWrapper.createdAt.Add(tcpStreamChannelTimeout)) {
					stream.Close()
					appStats.IncDroppedTcpStreams()
					rlog.Debugf("Dropped an unidentified TCP stream because of timeout. Total dropped: %d Total Goroutines: %d Timeout (ms): %d\n",
						appStats.DroppedTcpStreams, runtime.NumGoroutine(), tcpStreamChannelTimeout/1000000)
				}
			} else {
				if !stream.superIdentifier.IsClosedOthers {
					for i := range stream.clients {
						reader := &stream.clients[i]
						if reader.extension.Protocol != stream.superIdentifier.Protocol {
							reader.Close()
						}
					}
					for i := range stream.servers {
						reader := &stream.servers[i]
						if reader.extension.Protocol != stream.superIdentifier.Protocol {
							reader.Close()
						}
					}
					stream.superIdentifier.IsClosedOthers = true
				}
			}
			return true
		})
	}
}
