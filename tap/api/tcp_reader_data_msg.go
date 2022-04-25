package api

import "time"

type TcpReaderDataMsg interface {
	GetBytes() []byte
	GetTimestamp() time.Time
}

type tcpReaderDataMsg struct {
	bytes     []byte
	timestamp time.Time
}

func NewTcpReaderDataMsg(data []byte, timestamp time.Time) TcpReaderDataMsg {
	return &tcpReaderDataMsg{data, timestamp}
}

func (dataMsg *tcpReaderDataMsg) GetBytes() []byte {
	return dataMsg.bytes
}

func (dataMsg *tcpReaderDataMsg) GetTimestamp() time.Time {
	return dataMsg.timestamp
}
