package tlstapper

import (
	"bytes"
	"encoding/binary"
	"net"

	"github.com/go-errors/errors"
)

const FLAGS_IS_CLIENT_BIT uint32 = (1 << 0)
const FLAGS_IS_READ_BIT uint32 = (1 << 1)

// The same struct can be found in maps.h
//
//	Be careful when editing, alignment and padding should be exactly the same in go/c.
//
type tlsChunk struct {
	Pid      uint32
	Tgid     uint32
	Len      uint32
	Recorded uint32
	Fd       uint32
	Flags    uint32
	Address  [16]byte
	Data     [4096]byte
}

func (c *tlsChunk) getAddress() (net.IP, uint16, error) {
	address := bytes.NewReader(c.Address[:])
	var family uint16
	var port uint16
	var ip32 uint32

	if err := binary.Read(address, binary.BigEndian, &family); err != nil {
		return nil, 0, errors.Wrap(err, 0)
	}

	if err := binary.Read(address, binary.BigEndian, &port); err != nil {
		return nil, 0, errors.Wrap(err, 0)
	}

	if err := binary.Read(address, binary.BigEndian, &ip32); err != nil {
		return nil, 0, errors.Wrap(err, 0)
	}

	ip := net.IP{uint8(ip32 >> 24), uint8(ip32 >> 16), uint8(ip32 >> 8), uint8(ip32)}

	return ip, port, nil
}

func (c *tlsChunk) isClient() bool {
	return c.Flags&FLAGS_IS_CLIENT_BIT != 0
}

func (c *tlsChunk) isServer() bool {
	return !c.isClient()
}

func (c *tlsChunk) isRead() bool {
	return c.Flags&FLAGS_IS_READ_BIT != 0
}

func (c *tlsChunk) isWrite() bool {
	return !c.isRead()
}

func (c *tlsChunk) getRecordedData() []byte {
	return c.Data[:c.Recorded]
}

func (c *tlsChunk) isRequest() bool {
	return (c.isClient() && c.isWrite()) || (c.isServer() && c.isRead())
}
