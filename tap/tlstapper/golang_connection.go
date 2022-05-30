package tlstapper

type golangConnection struct {
	Pid         uint32
	ConnAddr    uint32
	AddressPair addressPair
	Request     []byte
	Response    []byte
	GotRequest  bool
	GotResponse bool
}

func NewGolangConnection(pid uint32, connAddr uint32) *golangConnection {
	return &golangConnection{
		Pid:      pid,
		ConnAddr: connAddr,
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
