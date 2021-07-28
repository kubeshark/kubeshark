package tap

import (
	"github.com/bradleyfalzon/tlsx"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func ExtractServerNameFromTLSClientHelloPacket(tcpPayload []byte) (string, error) {
	clientHello := tlsx.ClientHello{}
	err := clientHello.Unmarshall(tcpPayload)
	if err != nil {
		return "", err
	}
	return clientHello.SNI, nil
}

func IsTLSHandshakePacket(tcpPayload []byte) bool {
	tlsPacket := gopacket.NewPacket(tcpPayload, layers.LayerTypeTLS, gopacket.Default)

	if tlsPacket != nil && tlsPacket.Layer(layers.LayerTypeTLS) != nil {
		tlsLayer := tlsPacket.Layer(layers.LayerTypeTLS).(*layers.TLS)
		if tlsLayer != nil && len(tlsLayer.Handshake) > 0 && tlsLayer.Handshake[0].ContentType == layers.TLSHandshake {
			return true
		}
	}
	return false
}
