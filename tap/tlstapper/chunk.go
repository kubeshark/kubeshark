package tlstapper

import (
	"encoding/binary"
	"net"
	"unsafe"

	"github.com/up9inc/mizu/tap/api"
)

const FLAGS_IS_CLIENT_BIT uint32 = (1 << 0)
const FLAGS_IS_READ_BIT uint32 = (1 << 1)
const (
	addressInfoModeUndefined = iota
	addressInfoModeSingle
	addressInfoModePair
)

func (c *tlsTapperTlsChunk) getFdAddress() (net.IP, uint16, error) {
	sIP, sPort := c.getSrcAddress()
	return sIP, sPort, nil
}

func (c *tlsTapperTlsChunk) getDstAddress() (net.IP, uint16) {
	ip := intToIP(c.AddressInfo.Daddr)
	port := ntohs(c.AddressInfo.Dport)

	return ip, port
}

func (c *tlsTapperTlsChunk) getSrcAddress() (net.IP, uint16) {
	ip := intToIP(c.AddressInfo.Saddr)
	port := ntohs(c.AddressInfo.Sport)

	return ip, port
}

func (c *tlsTapperTlsChunk) getIsAddressPairValid() bool {
	if c.AddressInfo.Mode == addressInfoModePair {
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

func (c *tlsTapperTlsChunk) getKprobeAddressPair() (addressPair, bool) {
	dIP, dPort := c.getDstAddress()
	sIP, sPort := c.getSrcAddress()
	isAddressPairValid := c.getIsAddressPairValid()

	if c.isRequest() {
		return addressPair{
			srcIp:   sIP,
			srcPort: sPort,
			dstIp:   dIP,
			dstPort: dPort,
		},
		isAddressPairValid
	} else {
		return addressPair{
			srcIp:   dIP,
			srcPort: dPort,
			dstIp:   sIP,
			dstPort: sPort,
		},
		isAddressPairValid
	}
}

// intToIP converts IPv4 number to net.IP
func intToIP(ip32be uint32) net.IP {
	return net.IPv4(uint8(ip32be), uint8(ip32be >> 8), uint8(ip32be >> 16), uint8(ip32be >> 24))
}

// ntohs converts big endian (network byte order) to little endian (assuming that's the host byte order)
func ntohs(i16be uint16) uint16 {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, i16be)
	return *(*uint16)(unsafe.Pointer(&b[0]))
}
