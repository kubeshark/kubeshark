package tlstapper

import (
	"encoding/binary"
	"net"
	"unsafe"

	"github.com/up9inc/mizu/tap/api"
)

const FlagsIsClientBit uint32 = 1 << 0
const FlagsIsReadBit uint32 = 1 << 1
const (
	addressInfoModeUndefined = iota
	addressInfoModeSingle
	addressInfoModePair
)

func (c *tlsTapperTlsChunk) getSrcAddress() (net.IP, uint16) {
	ip := intToIP(c.AddressInfo.Saddr)
	port := ntohs(c.AddressInfo.Sport)

	return ip, port
}

func (c *tlsTapperTlsChunk) getDstAddress() (net.IP, uint16) {
	ip := intToIP(c.AddressInfo.Daddr)
	port := ntohs(c.AddressInfo.Dport)

	return ip, port
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

func (c *tlsTapperTlsChunk) getAddressPair() (addressPair, bool) {
	var (
		srcIp, dstIp     net.IP
		srcPort, dstPort uint16
		full             bool
	)

	switch c.AddressInfo.Mode {
	case addressInfoModeSingle:
		if c.isRequest() {
			srcIp, srcPort = api.UnknownIp, api.UnknownPort
			dstIp, dstPort = c.getSrcAddress()
		} else {
			srcIp, srcPort = c.getSrcAddress()
			dstIp, dstPort = api.UnknownIp, api.UnknownPort
		}
		full = false
	case addressInfoModePair:
		if c.isRequest() {
			srcIp, srcPort = c.getSrcAddress()
			dstIp, dstPort = c.getDstAddress()
		} else {
			srcIp, srcPort = c.getDstAddress()
			dstIp, dstPort = c.getSrcAddress()
		}
		full = true
	case addressInfoModeUndefined:
		srcIp, srcPort = api.UnknownIp, api.UnknownPort
		dstIp, dstPort = api.UnknownIp, api.UnknownPort
		full = false
	}

	return addressPair{
		srcIp:   srcIp,
		srcPort: srcPort,
		dstIp:   dstIp,
		dstPort: dstPort,
	}, full
}

// intToIP converts IPv4 number to net.IP
func intToIP(ip32be uint32) net.IP {
	return net.IPv4(uint8(ip32be), uint8(ip32be>>8), uint8(ip32be>>16), uint8(ip32be>>24))
}

// ntohs converts big endian (network byte order) to little endian (assuming that's the host byte order)
func ntohs(i16be uint16) uint16 {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, i16be)
	return *(*uint16)(unsafe.Pointer(&b[0]))
}
