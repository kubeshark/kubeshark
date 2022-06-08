package tlstapper

import (
	"github.com/cilium/ebpf/link"
	"github.com/go-errors/errors"
)

type golangHooks struct {
	golangWriteProbe    link.Link
	golangWriteExProbes []link.Link
	golangReadProbe     link.Link
	golangReadExProbes  []link.Link
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

	// Symbol points to
	// [`crypto/tls.(*Conn).Write`](https://github.com/golang/go/blob/go1.17.6/src/crypto/tls/conn.go#L1099)
	s.golangWriteProbe, err = ex.Uprobe(golangWriteSymbol, bpfObjects.GolangCryptoTlsWriteUprobe, &link.UprobeOptions{
		Offset: offsets.GolangWriteOffset.enter,
	})

	if err != nil {
		return errors.Wrap(err, 0)
	}

	for _, offset := range offsets.GolangWriteOffset.exits {
		probe, err := ex.Uprobe(golangWriteSymbol, bpfObjects.GolangCryptoTlsWriteExUprobe, &link.UprobeOptions{
			Offset: offset,
		})

		if err != nil {
			return errors.Wrap(err, 0)
		}

		s.golangWriteExProbes = append(s.golangWriteExProbes, probe)
	}

	// Symbol points to
	// [`crypto/tls.(*Conn).Read`](https://github.com/golang/go/blob/go1.17.6/src/crypto/tls/conn.go#L1263)
	s.golangReadProbe, err = ex.Uprobe(golangReadSymbol, bpfObjects.GolangCryptoTlsReadUprobe, &link.UprobeOptions{
		Offset: offsets.GolangReadOffset.enter,
	})

	if err != nil {
		return errors.Wrap(err, 0)
	}

	for _, offset := range offsets.GolangReadOffset.exits {
		probe, err := ex.Uprobe(golangReadSymbol, bpfObjects.GolangCryptoTlsReadExUprobe, &link.UprobeOptions{
			Offset: offset,
		})

		if err != nil {
			return errors.Wrap(err, 0)
		}

		s.golangReadExProbes = append(s.golangReadExProbes, probe)
	}

	return nil
}

func (s *golangHooks) close() []error {
	errors := make([]error, 0)

	if err := s.golangWriteProbe.Close(); err != nil {
		errors = append(errors, err)
	}

	for _, probe := range s.golangWriteExProbes {
		if err := probe.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if err := s.golangReadProbe.Close(); err != nil {
		errors = append(errors, err)
	}

	for _, probe := range s.golangReadExProbes {
		if err := probe.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}
