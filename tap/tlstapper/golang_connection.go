package tlstapper

import "github.com/up9inc/mizu/tap/api"

type golangConnection struct {
	pid          uint32
	connAddr     uint32
	addressPair  addressPair
	stream       *tlsStream
	clientReader *golangReader
	serverReader *golangReader
}

func NewGolangConnection(pid uint32, connAddr uint32, extension *api.Extension, emitter api.Emitter) *golangConnection {
	stream := &tlsStream{}
	counterPair := &api.CounterPair{}
	reqResMatcher := extension.Dissector.NewResponseRequestMatcher()
	clientReader := NewGolangReader(extension, true, emitter, counterPair, stream, reqResMatcher)
	serverReader := NewGolangReader(extension, false, emitter, counterPair, stream, reqResMatcher)
	stream.reader = clientReader
	return &golangConnection{
		pid:          pid,
		connAddr:     connAddr,
		stream:       stream,
		clientReader: clientReader,
		serverReader: serverReader,
	}
}

func (c *golangConnection) setAddressBySockfd(procfs string, pid uint32, fd uint32) error {
	addrPair, err := getAddressBySockfd(procfs, pid, fd)
	if err != nil {
		return err
	}
	c.addressPair = addrPair
	return nil
}
