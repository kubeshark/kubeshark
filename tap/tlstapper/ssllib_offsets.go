package tlstapper

import (
	"debug/elf"
	"fmt"

	"github.com/go-errors/errors"
	"github.com/up9inc/mizu/shared/logger"
)

type sslOffsets struct {
	SslWriteOffset   uint64
	SslReadOffset    uint64
	SslWriteExOffset uint64
	SslReadExOffset  uint64
}

func getSslOffsets(sslLibraryPath string) (sslOffsets, error) {
	sslElf, err := elf.Open(sslLibraryPath)

	if err != nil {
		return sslOffsets{}, errors.Wrap(err, 0)
	}

	defer sslElf.Close()

	base, err := findBaseAddress(sslElf, sslLibraryPath)

	if err != nil {
		return sslOffsets{}, errors.Wrap(err, 0)
	}

	offsets, err := findSslOffsets(sslElf, base)

	if err != nil {
		return sslOffsets{}, errors.Wrap(err, 0)
	}

	logger.Log.Debugf("Found TLS offsets (base: 0x%X) (write: 0x%X) (read: 0x%X)", base, offsets.SslWriteOffset, offsets.SslReadOffset)
	return offsets, nil
}

func findBaseAddress(sslElf *elf.File, sslLibraryPath string) (uint64, error) {
	for _, prog := range sslElf.Progs {
		if prog.Type == elf.PT_LOAD {
			return prog.Paddr, nil
		}
	}

	return 0, errors.New(fmt.Sprintf("Program header not found in %v", sslLibraryPath))
}

func findSslOffsets(sslElf *elf.File, base uint64) (sslOffsets, error) {
	symbolsMap := make(map[string]elf.Symbol)

	if err := buildSymbolsMap(sslElf.Symbols, &symbolsMap); err != nil {
		return sslOffsets{}, errors.Wrap(err, 0)
	}

	if err := buildSymbolsMap(sslElf.DynamicSymbols, &symbolsMap); err != nil {
		return sslOffsets{}, errors.Wrap(err, 0)
	}

	var sslWriteOffset, sslReadOffset, sslWriteExOffset, sslReadExOffset elf.Symbol
	var ok bool

	if sslWriteOffset, ok = symbolsMap["SSL_write"]; !ok {
		return sslOffsets{}, errors.New("SSL_write symbol not found")
	}

	if sslReadOffset, ok = symbolsMap["SSL_read"]; !ok {
		return sslOffsets{}, errors.New("SSL_read symbol not found")
	}

	if sslWriteExOffset, ok = symbolsMap["SSL_write_ex"]; !ok {
		return sslOffsets{}, errors.New("SSL_write_ex symbol not found")
	}

	if sslReadExOffset, ok = symbolsMap["SSL_read_ex"]; !ok {
		return sslOffsets{}, errors.New("SSL_read_ex symbol not found")
	}

	return sslOffsets{
		SslWriteOffset:   sslWriteOffset.Value - base,
		SslReadOffset:    sslReadOffset.Value - base,
		SslWriteExOffset: sslWriteExOffset.Value - base,
		SslReadExOffset:  sslReadExOffset.Value - base,
	}, nil
}

func buildSymbolsMap(sectionGetter func() ([]elf.Symbol, error), symbols *map[string]elf.Symbol) error {
	syms, err := sectionGetter()

	if err != nil && !errors.Is(err, elf.ErrNoSymbols) {
		return err
	}

	for _, sym := range syms {
		if elf.ST_TYPE(sym.Info) != elf.STT_FUNC {
			continue
		}

		(*symbols)[sym.Name] = sym
	}

	return nil
}
