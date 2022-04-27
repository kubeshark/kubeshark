package tap

import (
	"os"
	"runtime"
	_debug "runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/tap/diagnose"
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

func (streamMap *tcpStreamMap) getCloseTimedoutTcpChannelsInterval() time.Duration {
	defaultDuration := 1000 * time.Millisecond
	rangeMin := 10
	rangeMax := 10000
	closeTimedoutTcpChannelsIntervalMsStr := os.Getenv(CloseTimedoutTcpChannelsIntervalMsEnvVar)
	if closeTimedoutTcpChannelsIntervalMsStr == "" {
		return defaultDuration
	} else {
		closeTimedoutTcpChannelsIntervalMs, err := strconv.Atoi(closeTimedoutTcpChannelsIntervalMsStr)
		if err != nil {
			logger.Log.Warningf("Error parsing environment variable %s: %v\n", CloseTimedoutTcpChannelsIntervalMsEnvVar, err)
			return defaultDuration
		} else {
			if closeTimedoutTcpChannelsIntervalMs < rangeMin || closeTimedoutTcpChannelsIntervalMs > rangeMax {
				logger.Log.Warningf("The value of environment variable %s is not in acceptable range: %d - %d\n", CloseTimedoutTcpChannelsIntervalMsEnvVar, rangeMin, rangeMax)
				return defaultDuration
			} else {
				return time.Duration(closeTimedoutTcpChannelsIntervalMs) * time.Millisecond
			}
		}
	}
}

func (streamMap *tcpStreamMap) closeTimedoutTcpStreamChannels() {
	tcpStreamChannelTimeout := GetTcpChannelTimeoutMs()
	closeTimedoutTcpChannelsIntervalMs := streamMap.getCloseTimedoutTcpChannelsInterval()
	logger.Log.Infof("Using %d ms as the close timedout TCP stream channels interval", closeTimedoutTcpChannelsIntervalMs/time.Millisecond)
	for {
		time.Sleep(closeTimedoutTcpChannelsIntervalMs)
		_debug.FreeOSMemory()
		streamMap.streams.Range(func(key interface{}, value interface{}) bool {
			streamWrapper := value.(*tcpStreamWrapper)
			stream := streamWrapper.stream
			if stream.superIdentifier.Protocol == nil {
				if !stream.isClosed && time.Now().After(streamWrapper.createdAt.Add(tcpStreamChannelTimeout)) {
					stream.Close()
					diagnose.AppStats.IncDroppedTcpStreams()
					logger.Log.Debugf("Dropped an unidentified TCP stream because of timeout. Total dropped: %d Total Goroutines: %d Timeout (ms): %d",
						diagnose.AppStats.DroppedTcpStreams, runtime.NumGoroutine(), tcpStreamChannelTimeout/time.Millisecond)
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
