package tap

import (
	"encoding/binary"
)

func IsPacketTLSClientHello(tcpPayload []byte) bool {
	if tcpPayload != nil && len(tcpPayload) > 200 {
		if tcpPayload[0] == 22 && tcpPayload[1] == 3 && tcpPayload[2] == 1 && tcpPayload[5] == 1{
			recordSize := binary.BigEndian.Uint16([]byte {tcpPayload[3], tcpPayload[4]})
			if recordSize == uint16(len(tcpPayload) - 5) {
				handShakeSize := binary.BigEndian.Uint16([]byte {tcpPayload[7], tcpPayload[8]})
				if handShakeSize == uint16(len(tcpPayload) - 9) {
					return true
				}
			}
		}
	}
	return false
}

func ExtractServerNameFromTLSClientHello(tcpPayload []byte)  {
	//todo: iterate over packet bytes, checking dynamic length parts until you react the
}
