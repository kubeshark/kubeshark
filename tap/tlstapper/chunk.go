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
	Pid      uint32     // process id
	Tgid     uint32     // thread id inside the process
	Len      uint32     // the size of the native buffer used to read/write the tls data (may be bigger than tlsChunk.Data[])
	Start    uint32     // the start offset withing the native buffer
	Recorded uint32     // number of bytes copied from the native buffer to tlsChunk.Data[]
	Fd       uint32     // the file descriptor used to read/write the tls data (probably socket file descriptor)
	Flags    uint32     // bitwise flags
	Address  [16]byte   // ipv4 address and port
	Data     [4096]byte // actual tls data
}

func (c *tlsChunk) GetAddress() (net.IP, uint16, error) {
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

func (c *tlsChunk) IsClient() bool {
	return c.Flags&FLAGS_IS_CLIENT_BIT != 0
}

func (c *tlsChunk) IsServer() bool {
	return !c.IsClient()
}

func (c *tlsChunk) IsRead() bool {
	return c.Flags&FLAGS_IS_READ_BIT != 0
}

func (c *tlsChunk) IsWrite() bool {
	return !c.IsRead()
}

func (c *tlsChunk) GetRecordedData() []byte {
	return c.Data[:c.Recorded]
}

func (c *tlsChunk) IsRequest() bool {
	return (c.IsClient() && c.IsWrite()) || (c.IsServer() && c.IsRead())
}
