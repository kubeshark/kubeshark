package tlstapper

import (
	"github.com/cilium/ebpf/link"
	"github.com/go-errors/errors"
)

type sslHooks struct {
	sslWriteProbe      link.Link
	sslWriteRetProbe   link.Link
	sslReadProbe       link.Link
	sslReadRetProbe    link.Link
	sslWriteExProbe    link.Link
	sslWriteExRetProbe link.Link
	sslReadExProbe     link.Link
	sslReadExRetProbe  link.Link
}

func (s *sslHooks) installUprobes(bpfObjects *tlsTapperObjects, sslLibraryPath string) error {
	sslLibrary, err := link.OpenExecutable(sslLibraryPath)

	if err != nil {
		return errors.Wrap(err, 0)
	}

	sslOffsets, err := getSslOffsets(sslLibraryPath)

	if err != nil {
		return errors.Wrap(err, 0)
	}

	return s.installSslHooks(bpfObjects, sslLibrary, sslOffsets)
}

func (s *sslHooks) installSslHooks(bpfObjects *tlsTapperObjects, sslLibrary *link.Executable, offsets sslOffsets) error {
	var err error

	s.sslWriteProbe, err = sslLibrary.Uprobe("SSL_write", bpfObjects.SslWrite, &link.UprobeOptions{
		Offset: offsets.SslWriteOffset,
	})

	if err != nil {
		return errors.Wrap(err, 0)
	}

	s.sslWriteRetProbe, err = sslLibrary.Uretprobe("SSL_write", bpfObjects.SslRetWrite, &link.UprobeOptions{
		Offset: offsets.SslWriteOffset,
	})

	if err != nil {
		return errors.Wrap(err, 0)
	}

	s.sslReadProbe, err = sslLibrary.Uprobe("SSL_read", bpfObjects.SslRead, &link.UprobeOptions{
		Offset: offsets.SslReadOffset,
	})

	if err != nil {
		return errors.Wrap(err, 0)
	}

	s.sslReadRetProbe, err = sslLibrary.Uretprobe("SSL_read", bpfObjects.SslRetRead, &link.UprobeOptions{
		Offset: offsets.SslReadOffset,
	})

	if err != nil {
		return errors.Wrap(err, 0)
	}

	if offsets.SslWriteExOffset != 0 {
		s.sslWriteExProbe, err = sslLibrary.Uprobe("SSL_write_ex", bpfObjects.SslWriteEx, &link.UprobeOptions{
			Offset: offsets.SslWriteExOffset,
		})

		if err != nil {
			return errors.Wrap(err, 0)
		}

		s.sslWriteExRetProbe, err = sslLibrary.Uretprobe("SSL_write_ex", bpfObjects.SslRetWriteEx, &link.UprobeOptions{
			Offset: offsets.SslWriteExOffset,
		})

		if err != nil {
			return errors.Wrap(err, 0)
		}
	}

	if offsets.SslReadExOffset != 0 {
		s.sslReadExProbe, err = sslLibrary.Uprobe("SSL_read_ex", bpfObjects.SslReadEx, &link.UprobeOptions{
			Offset: offsets.SslReadExOffset,
		})

		if err != nil {
			return errors.Wrap(err, 0)
		}

		s.sslReadExRetProbe, err = sslLibrary.Uretprobe("SSL_read_ex", bpfObjects.SslRetReadEx, &link.UprobeOptions{
			Offset: offsets.SslReadExOffset,
		})

		if err != nil {
			return errors.Wrap(err, 0)
		}
	}

	return nil
}

func (s *sslHooks) close() []error {
	errors := make([]error, 0)

	if err := s.sslWriteProbe.Close(); err != nil {
		errors = append(errors, err)
	}

	if err := s.sslWriteRetProbe.Close(); err != nil {
		errors = append(errors, err)
	}

	if err := s.sslReadProbe.Close(); err != nil {
		errors = append(errors, err)
	}

	if err := s.sslReadRetProbe.Close(); err != nil {
		errors = append(errors, err)
	}

	if s.sslWriteExProbe != nil {
		if err := s.sslWriteExProbe.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if s.sslWriteExRetProbe != nil {
		if err := s.sslWriteExRetProbe.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if s.sslReadExProbe != nil {
		if err := s.sslReadExProbe.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if s.sslReadExRetProbe != nil {
		if err := s.sslReadExRetProbe.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}
