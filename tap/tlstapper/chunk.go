package tlstapper

import (
	"bytes"
	"encoding/binary"
	"net"

	"github.com/go-errors/errors"
	"github.com/up9inc/mizu/tap/api"
)

const FlagsIsClientBit uint32 = 1 << 0
const FlagsIsReadBit uint32 = 1 << 1

func (c *tlsTapperTlsChunk) getAddress() (net.IP, uint16, error) {
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

func (c *tlsTapperTlsChunk) isClient() bool {
	return c.Flags&FlagsIsClientBit != 0
}

func (c *tlsTapperTlsChunk) isServer() bool {
	return !c.isClient()
}

func (c *tlsTapperTlsChunk) isRead() bool {
	return c.Flags&FlagsIsReadBit != 0
}

func (c *tlsTapperTlsChunk) isWrite() bool {
	return !c.isRead()
}

func (c *tlsTapperTlsChunk) getRecordedData() []byte {
	return c.Data[:c.Recorded]
}

func (c *tlsTapperTlsChunk) isRequest() bool {
	return (c.isClient() && c.isWrite()) || (c.isServer() && c.isRead())
}

func (c *tlsTapperTlsChunk) getAddressPair() (addressPair, error) {
	ip, port, err := c.getAddress()

	if err != nil {
		return addressPair{}, err
	}

	if c.isRequest() {
		return addressPair{
			srcIp:   api.UnknownIp,
			srcPort: api.UnknownPort,
			dstIp:   ip,
			dstPort: port,
		}, nil
	} else {
		return addressPair{
			srcIp:   ip,
			srcPort: port,
			dstIp:   api.UnknownIp,
			dstPort: api.UnknownPort,
		}, nil
	}
}
