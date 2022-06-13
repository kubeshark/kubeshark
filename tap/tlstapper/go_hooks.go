package tlstapper

import (
	"github.com/cilium/ebpf/link"
	"github.com/go-errors/errors"
)

type goHooks struct {
	goWriteProbe    link.Link
	goWriteExProbes []link.Link
	goReadProbe     link.Link
	goReadExProbes  []link.Link
}

func (s *goHooks) installUprobes(bpfObjects *tlsTapperObjects, filePath string) error {
	ex, err := link.OpenExecutable(filePath)

	if err != nil {
		return errors.Wrap(err, 0)
	}

	offsets, err := findGoOffsets(filePath)

	if err != nil {
		return errors.Wrap(err, 0)
	}

	return s.installHooks(bpfObjects, ex, offsets)
}

func (s *goHooks) installHooks(bpfObjects *tlsTapperObjects, ex *link.Executable, offsets goOffsets) error {
	var err error

	// Symbol points to
	// [`crypto/tls.(*Conn).Write`](https://github.com/golang/go/blob/go1.17.6/src/crypto/tls/conn.go#L1099)
	s.goWriteProbe, err = ex.Uprobe(goWriteSymbol, bpfObjects.GoCryptoTlsWrite, &link.UprobeOptions{
		Offset: offsets.GoWriteOffset.enter,
	})

	if err != nil {
		return errors.Wrap(err, 0)
	}

	for _, offset := range offsets.GoWriteOffset.exits {
		probe, err := ex.Uprobe(goWriteSymbol, bpfObjects.GoCryptoTlsWriteEx, &link.UprobeOptions{
			Offset: offset,
		})

		if err != nil {
			return errors.Wrap(err, 0)
		}

		s.goWriteExProbes = append(s.goWriteExProbes, probe)
	}

	// Symbol points to
	// [`crypto/tls.(*Conn).Read`](https://github.com/golang/go/blob/go1.17.6/src/crypto/tls/conn.go#L1263)
	s.goReadProbe, err = ex.Uprobe(goReadSymbol, bpfObjects.GoCryptoTlsRead, &link.UprobeOptions{
		Offset: offsets.GoReadOffset.enter,
	})

	if err != nil {
		return errors.Wrap(err, 0)
	}

	for _, offset := range offsets.GoReadOffset.exits {
		probe, err := ex.Uprobe(goReadSymbol, bpfObjects.GoCryptoTlsReadEx, &link.UprobeOptions{
			Offset: offset,
		})

		if err != nil {
			return errors.Wrap(err, 0)
		}

		s.goReadExProbes = append(s.goReadExProbes, probe)
	}

	return nil
}

func (s *goHooks) close() []error {
	errors := make([]error, 0)

	if err := s.goWriteProbe.Close(); err != nil {
		errors = append(errors, err)
	}

	for _, probe := range s.goWriteExProbes {
		if err := probe.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if err := s.goReadProbe.Close(); err != nil {
		errors = append(errors, err)
	}

	for _, probe := range s.goReadExProbes {
		if err := probe.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}
