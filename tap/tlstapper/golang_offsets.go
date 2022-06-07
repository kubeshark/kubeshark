package tlstapper

import (
	"bufio"
	"debug/elf"
	"fmt"
	"os"

	"github.com/Masterminds/semver"
	"github.com/cilium/ebpf/link"
)

type golangOffsets struct {
	GolangWriteOffset uint64
	GolangReadOffset  uint64
}

const (
	minimumSupportedGoVersion = "1.17.0"
	golangVersionSymbol       = "runtime.buildVersion.str"
	golangWriteSymbol         = "crypto/tls.(*Conn).Write"
	golangReadSymbol          = "crypto/tls.(*Conn).Read"
)

func findGolangOffsets(filePath string) (golangOffsets, error) {
	offsets, err := getOffsets(filePath)
	if err != nil {
		return golangOffsets{}, err
	}

	goVersionOffset, err := getOffset(offsets, golangVersionSymbol)
	if err != nil {
		return golangOffsets{}, err
	}

	passed, goVersion, err := checkGoVersion(filePath, goVersionOffset)
	if err != nil {
		return golangOffsets{}, fmt.Errorf("Checking Go version: %s", err)
	}

	if !passed {
		return golangOffsets{}, fmt.Errorf("Unsupported Go version: %s", goVersion)
	}

	writeOffset, err := getOffset(offsets, golangWriteSymbol)
	if err != nil {
		return golangOffsets{}, fmt.Errorf("reading offset [%s]: %s", golangWriteSymbol, err)
	}

	readOffset, err := getOffset(offsets, golangReadSymbol)
	if err != nil {
		return golangOffsets{}, fmt.Errorf("reading offset [%s]: %s", golangReadSymbol, err)
	}

	return golangOffsets{
		GolangWriteOffset: writeOffset,
		GolangReadOffset:  readOffset,
	}, nil
}

func getOffsets(filePath string) (offsets map[string]uint64, err error) {
	offsets = make(map[string]uint64)
	var fd *os.File
	fd, err = os.Open(filePath)
	if err != nil {
		return
	}
	defer fd.Close()

	var se *elf.File
	se, err = elf.NewFile(fd)
	if err != nil {
		return nil, err
	}

	syms, err := se.Symbols()
	for _, sym := range syms {
		offset := sym.Value

		for _, prog := range se.Progs {
			if prog.Vaddr <= sym.Value && sym.Value < (prog.Vaddr+prog.Memsz) {
				offset = sym.Value - prog.Vaddr + prog.Off
				break
			}
		}

		offsets[sym.Name] = offset
	}

	return
}

func getOffset(offsets map[string]uint64, symbol string) (uint64, error) {
	if offset, ok := offsets[symbol]; ok {
		return offset, nil
	}
	return 0, fmt.Errorf("symbol %s: %w", symbol, link.ErrNoSymbol)
}

func checkGoVersion(filePath string, offset uint64) (bool, string, error) {
	fd, err := os.Open(filePath)
	if err != nil {
		return false, "", err
	}
	defer fd.Close()

	reader := bufio.NewReader(fd)

	_, err = reader.Discard(int(offset))
	if err != nil {
		return false, "", err
	}

	line, err := reader.ReadString(0)
	if err != nil {
		return false, "", err
	}

	if len(line) < 3 {
		return false, "", fmt.Errorf("ELF data segment read error (corrupted result)")
	}

	goVersionStr := line[2 : len(line)-1]

	goVersion, err := semver.NewVersion(goVersionStr)
	if err != nil {
		return false, goVersionStr, err
	}

	goVersionConstraint, err := semver.NewConstraint(fmt.Sprintf(">= %s", minimumSupportedGoVersion))
	if err != nil {
		return false, goVersionStr, err
	}

	return goVersionConstraint.Check(goVersion), goVersionStr, nil
}
