package tlstapper

import "github.com/up9inc/mizu/tap/api"

type golangConnection struct {
	Pid          uint32
	ConnAddr     uint32
	AddressPair  addressPair
	Requests     [][]byte
	Responses    [][]byte
	Gzipped      bool
	Stream       *tlsStream
	ClientReader *golangReader
	ServerReader *golangReader
}

func NewGolangConnection(pid uint32, connAddr uint32, extension *api.Extension, emitter api.Emitter) *golangConnection {
	stream := &tlsStream{}
	return &golangConnection{
		Pid:          pid,
		ConnAddr:     connAddr,
		Stream:       stream,
		ClientReader: NewGolangReader(extension, emitter, stream, true),
		ServerReader: NewGolangReader(extension, emitter, stream, false),
	}
}

func (c *golangConnection) setAddressBySockfd(procfs string, pid uint32, fd uint32) error {
	addrPair, err := getAddressBySockfd(procfs, pid, fd)
	if err != nil {
		return err
	}
	c.AddressPair = addrPair
	return nil
}
