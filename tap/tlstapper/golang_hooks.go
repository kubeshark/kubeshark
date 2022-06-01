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
	golangGzipProbe   link.Link
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
	// [`net/http.(*Transport).dialConn+412`](https://github.com/golang/go/blob/go1.17.6/src/net/http/transport.go#L1561)
	s.golangDialProbe, err = ex.Uprobe(golangDialSymbol, bpfObjects.GolangNetHttpDialconnUprobe, &link.UprobeOptions{
		Offset: offsets.GolangDialOffset + 0x19c,
	})

	if err != nil {
		return errors.Wrap(err, 0)
	}

	// Relative offset points to
	// [`net.socket+127`](https://github.com/golang/go/blob/go1.17.6/src/net/sock_posix.go#L24)
	s.golangSocketProbe, err = ex.Uprobe(golangSocketSymbol, bpfObjects.GolangNetSocketUprobe, &link.UprobeOptions{
		Offset: offsets.GolangSocketOffset + 0x7f,
	})

	if err != nil {
		return errors.Wrap(err, 0)
	}

	// Symbol points to
	// [`crypto/tls.(*Conn).Write`](https://github.com/golang/go/blob/go1.17.6/src/crypto/tls/conn.go#L1099)
	s.golangWriteProbe, err = ex.Uprobe(golangWriteSymbol, bpfObjects.GolangCryptoTlsWriteUprobe, &link.UprobeOptions{
		Offset: offsets.GolangWriteOffset,
	})

	if err != nil {
		return errors.Wrap(err, 0)
	}

	// Relative offset points to
	// [`crypto/tls.(*Conn).Read+559`](https://github.com/golang/go/blob/go1.17.6/src/crypto/tls/conn.go#L1296)
	s.golangReadProbe, err = ex.Uprobe(golangReadSymbol, bpfObjects.GolangCryptoTlsReadUprobe, &link.UprobeOptions{
		Offset: offsets.GolangReadOffset + 0x22f,
	})

	if err != nil {
		return errors.Wrap(err, 0)
	}

	// Relative offset points to
	// [`net/http.(*gzipReader).Read+363`](https://github.com/golang/go/blob/go1.17.6/src/net/http/transport.go#L2832)
	s.golangGzipProbe, err = ex.Uprobe(golangReadSymbol, bpfObjects.GolangNetHttpGzipreaderReadUprobe, &link.UprobeOptions{
		Offset: offsets.GolangGzipOffset + 0x16b,
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

	if err := s.golangGzipProbe.Close(); err != nil {
		errors = append(errors, err)
	}

	return errors
}
