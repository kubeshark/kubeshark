package tlstapper

import (
	"bufio"
	"debug/elf"
	"fmt"
	"os"

	"github.com/Masterminds/semver"
	"github.com/cilium/ebpf/link"
	"github.com/knightsc/gapstone"
)

type golangOffsets struct {
	GolangWriteOffset *golangExtendedOffset
	GolangReadOffset  *golangExtendedOffset
}

type golangExtendedOffset struct {
	enter uint64
	exits []uint64
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

func getOffsets(filePath string) (offsets map[string]*golangExtendedOffset, err error) {
	var engine gapstone.Engine
	engine, err = gapstone.New(
		gapstone.CS_ARCH_X86,
		gapstone.CS_MODE_64,
	)
	if err != nil {
		return
	}

	offsets = make(map[string]*golangExtendedOffset)
	var fd *os.File
	fd, err = os.Open(filePath)
	if err != nil {
		return
	}
	defer fd.Close()

	var se *elf.File
	se, err = elf.NewFile(fd)
	if err != nil {
		return
	}

	textSection := se.Section(".text")
	if textSection == nil {
		err = fmt.Errorf("No text section")
		return
	}

	// extract the raw bytes from the .text section
	var textSectionData []byte
	textSectionData, err = textSection.Data()
	if err != nil {
		return
	}

	syms, err := se.Symbols()
	for _, sym := range syms {
		offset := sym.Value

		var lastProg *elf.Prog
		for _, prog := range se.Progs {
			if prog.Vaddr <= sym.Value && sym.Value < (prog.Vaddr+prog.Memsz) {
				offset = sym.Value - prog.Vaddr + prog.Off
				lastProg = prog
				break
			}
		}

		extendedOffset := &golangExtendedOffset{enter: offset}

		// source: https://gist.github.com/grantseltzer/3efa8ecc5de1fb566e8091533050d608
		// skip over any symbols that aren't functinons/methods
		if sym.Info != byte(2) && sym.Info != byte(18) {
			offsets[sym.Name] = extendedOffset
			continue
		}

		// skip over empty symbols
		if sym.Size == 0 {
			offsets[sym.Name] = extendedOffset
			continue
		}

		// calculate starting and ending index of the symbol within the text section
		symStartingIndex := sym.Value - textSection.Addr
		symEndingIndex := symStartingIndex + sym.Size

		// collect the bytes of the symbol
		symBytes := textSectionData[symStartingIndex:symEndingIndex]

		// disasemble the symbol
		var instructions []gapstone.Instruction
		instructions, err = engine.Disasm(symBytes, sym.Value, 0)
		if err != nil {
			return
		}

		// iterate over each instruction and if the mnemonic is `ret` then that's an exit offset
		for _, ins := range instructions {
			if ins.Mnemonic == "ret" {
				extendedOffset.exits = append(extendedOffset.exits, uint64(ins.Address)-lastProg.Vaddr+lastProg.Off)
			}
		}

		offsets[sym.Name] = extendedOffset
	}

	return
}

func getOffset(offsets map[string]*golangExtendedOffset, symbol string) (*golangExtendedOffset, error) {
	if offset, ok := offsets[symbol]; ok {
		return offset, nil
	}
	return nil, fmt.Errorf("symbol %s: %w", symbol, link.ErrNoSymbol)
}

func checkGoVersion(filePath string, offset *golangExtendedOffset) (bool, string, error) {
	fd, err := os.Open(filePath)
	if err != nil {
		return false, "", err
	}
	defer fd.Close()

	reader := bufio.NewReader(fd)

	_, err = reader.Discard(int(offset.enter))
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
