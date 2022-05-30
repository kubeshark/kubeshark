package tlstapper

import (
	"github.com/cilium/ebpf/link"
	"github.com/go-errors/errors"
)

type golangHooks struct {
	golangDialProbe   link.Link
	golangSocketProbe link.Link
	golangWriteProbe  link.Link
	golangReadProbe   link.Link
}

func (s *golangHooks) installUprobes(bpfObjects *tlsTapperObjects, filePath string) error {
	ex, err := link.OpenExecutable(filePath)

	if err != nil {
		return errors.Wrap(err, 0)
	}

	offsets, err := findGolangOffsets(filePath)

	if err != nil {
		return errors.Wrap(err, 0)
	}

	return s.installHooks(bpfObjects, ex, offsets)
}

func (s *golangHooks) installHooks(bpfObjects *tlsTapperObjects, ex *link.Executable, offsets golangOffsets) error {
	var err error

	// Relative offset points to
	// [`net/http.(*Transport).dialConn+412`](https://github.com/golang/go/blob/fe4de36198794c447fbd9d7cc2d7199a506c76a5/src/net/http/transport.go#L1564)
	s.golangDialProbe, err = ex.Uprobe(golangDialSymbol, bpfObjects.GolangNetHttpDialconnUprobe, &link.UprobeOptions{
		Offset: offsets.GolangDialOffset + 0x19c,
	})

	if err != nil {
		return errors.Wrap(err, 0)
	}

	// Relative offset points to
	// [`net.socket+127`](https://github.com/golang/go/blob/fe4de36198794c447fbd9d7cc2d7199a506c76a5/src/net/sock_posix.go#L23)
	s.golangSocketProbe, err = ex.Uprobe(golangSocketSymbol, bpfObjects.GolangNetSocketUprobe, &link.UprobeOptions{
		Offset: offsets.GolangSocketOffset + 0x7f,
	})

	if err != nil {
		return errors.Wrap(err, 0)
	}

	// Symbol points to
	// [`crypto/tls.(*Conn).Write`](https://github.com/golang/go/blob/fe4de36198794c447fbd9d7cc2d7199a506c76a5/src/crypto/tls/conn.go#L1109)
	s.golangWriteProbe, err = ex.Uprobe(golangWriteSymbol, bpfObjects.GolangCryptoTlsWriteUprobe, &link.UprobeOptions{
		Offset: offsets.GolangWriteOffset,
	})

	if err != nil {
		return errors.Wrap(err, 0)
	}

	// Relative offset points to
	// [`net/http.(*persistConn).Read+92`](https://github.com/golang/go/blob/fe4de36198794c447fbd9d7cc2d7199a506c76a5/src/net/http/transport.go#L1929)
	s.golangReadProbe, err = ex.Uprobe(golangReadSymbol, bpfObjects.GolangNetHttpReadUprobe, &link.UprobeOptions{
		Offset: offsets.GolangReadOffset + 0x5c,
	})

	if err != nil {
		return errors.Wrap(err, 0)
	}

	return nil
}

func (s *golangHooks) close() []error {
	errors := make([]error, 0)

	if err := s.golangDialProbe.Close(); err != nil {
		errors = append(errors, err)
	}

	if err := s.golangSocketProbe.Close(); err != nil {
		errors = append(errors, err)
	}

	if err := s.golangWriteProbe.Close(); err != nil {
		errors = append(errors, err)
	}

	if err := s.golangReadProbe.Close(); err != nil {
		errors = append(errors, err)
	}

	return errors
}
