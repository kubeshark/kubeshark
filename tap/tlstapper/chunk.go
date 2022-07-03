package tlstapper

import (
	"bytes"
	"encoding/binary"
	"net"

	"github.com/go-errors/errors"
	"github.com/up9inc/mizu/tap/api"
)

const FLAGS_IS_CLIENT_BIT uint32 = (1 << 0)
const FLAGS_IS_READ_BIT uint32 = (1 << 1)

func (c *tlsTapperTlsChunk) getFdAddress() (net.IP, uint16, error) {
	address := bytes.NewReader(c.FdAddress[:])
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

func (c *tlsTapperTlsChunk) getDstAddress() (net.IP, uint16) {
	ip := intToIP(c.KprobeAddressPair.Daddr)
	port := c.KprobeAddressPair.Dport

	return ip, port
}

func (c *tlsTapperTlsChunk) getSrcAddress() (net.IP, uint16) {
	ip := intToIP(c.KprobeAddressPair.Saddr)
	port := c.KprobeAddressPair.Sport

	return ip, port
}

func (c *tlsTapperTlsChunk) getIsAddressPairValid() bool {
	if c.KprobeAddressPair.IsAddressPairValid == 1 {
		return true
	}

	return false
}

func (c *tlsTapperTlsChunk) isClient() bool {
	return c.Flags&FLAGS_IS_CLIENT_BIT != 0
}

func (c *tlsTapperTlsChunk) isServer() bool {
	return !c.isClient()
}

func (c *tlsTapperTlsChunk) isRead() bool {
	return c.Flags&FLAGS_IS_READ_BIT != 0
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

func (c *tlsTapperTlsChunk) getFdPartialAddressPair() (addressPair, error) {
	ip, port, err := c.getFdAddress()

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

func (c *tlsTapperTlsChunk) getKprobeAddressPair() (bool, addressPair) {
	dIP, dPort := c.getDstAddress()
	sIP, sPort := c.getSrcAddress()
	isAddressPairValid := c.getIsAddressPairValid()

	if c.isRequest() {
		return isAddressPairValid,
		addressPair{
			srcIp:   sIP,
			srcPort: sPort,
			dstIp:   dIP,
			dstPort: dPort,
		}
	} else {
		return isAddressPairValid,
		addressPair{
			srcIp:   dIP,
			srcPort: dPort,
			dstIp:   sIP,
			dstPort: sPort,
		}
	}
}

// intToIP converts IPv4 number to net.IP
func intToIP(ipNum uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, ipNum)
	return ip
}
