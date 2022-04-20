package api

import (
	"sync"
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
