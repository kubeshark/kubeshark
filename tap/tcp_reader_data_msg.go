package tap

import (
	"time"

	"github.com/up9inc/mizu/tap/api"
)

type tcpReaderDataMsg struct {
	bytes     []byte
	timestamp time.Time
}

func NewTcpReaderDataMsg(data []byte, timestamp time.Time) api.TcpReaderDataMsg {
	return &tcpReaderDataMsg{data, timestamp}
}

func (dataMsg *tcpReaderDataMsg) GetBytes() []byte {
	return dataMsg.bytes
}

func (dataMsg *tcpReaderDataMsg) GetTimestamp() time.Time {
	return dataMsg.timestamp
}
