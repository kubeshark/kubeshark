package tlstapper

import (
	"bufio"
	"debug/dwarf"
	"debug/elf"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/Masterminds/semver"
	"github.com/cilium/ebpf/link"
	"github.com/knightsc/gapstone"
	"github.com/kubeshark/kubeshark/logger"
)

type goAbi int

const (
	ABI0 goAbi = iota
	ABIInternal
)

const PtrSize int = 8

type goOffsets struct {
	GoWriteOffset *goExtendedOffset
	GoReadOffset  *goExtendedOffset
	GoVersion     string
	Abi           goAbi
	GoidOffset    uint64
	GStructOffset uint64
}

type goExtendedOffset struct {
	enter uint64
	exits []uint64
}

const (
	minimumABIInternalGoVersion = "1.17.0"
	goVersionSymbol             = "runtime.buildVersion.str" // symbol does not exist in Go (<=1.16)
	goWriteSymbol               = "crypto/tls.(*Conn).Write"
	goReadSymbol                = "crypto/tls.(*Conn).Read"
)

func findGoOffsets(filePath string) (goOffsets, error) {
	offsets, goidOffset, gStructOffset, err := getOffsets(filePath)
	if err != nil {
		return goOffsets{}, err
	}

	abi := ABI0
	var passed bool
	var goVersion string

	goVersionOffset, err := getOffset(offsets, goVersionSymbol)
	if err == nil {
		// TODO: Replace this logic with https://pkg.go.dev/debug/buildinfo#ReadFile once we upgrade to 1.18
		passed, goVersion, err = checkGoVersion(filePath, goVersionOffset)
		if err != nil {
			return goOffsets{}, fmt.Errorf("Checking Go version: %s", err)
		}
	}

	if passed {
		abi = ABIInternal
	}

	writeOffset, err := getOffset(offsets, goWriteSymbol)
	if err != nil {
		return goOffsets{}, fmt.Errorf("reading offset [%s]: %s", goWriteSymbol, err)
	}

	readOffset, err := getOffset(offsets, goReadSymbol)
	if err != nil {
		return goOffsets{}, fmt.Errorf("reading offset [%s]: %s", goReadSymbol, err)
	}

	return goOffsets{
		GoWriteOffset: writeOffset,
		GoReadOffset:  readOffset,
		GoVersion:     goVersion,
		Abi:           abi,
		GoidOffset:    goidOffset,
		GStructOffset: gStructOffset,
	}, nil
}

func getSymbol(exe *elf.File, name string) *elf.Symbol {
	symbols, err := exe.Symbols()
	if err != nil {
		return nil
	}

	for _, symbol := range symbols {
		if symbol.Name == name {
			s := symbol
			return &s
		}
	}
	return nil
}

func getGStructOffset(exe *elf.File) (gStructOffset uint64, err error) {
	// This is a bit arcane. Essentially:
	// - If the program is pure Go, it can do whatever it wants, and puts the G
	//   pointer at %fs-8 on 64 bit.
	// - %Gs is the index of private storage in GDT on 32 bit, and puts the G
	//   pointer at -4(tls).
	// - Otherwise, Go asks the external linker to place the G pointer by
	//   emitting runtime.tlsg, a TLS symbol, which is relocated to the chosen
	//   offset in libc's TLS block.
	// - On ARM64 (but really, any architecture other than i386 and 86x64) the
	//   offset is calculate using runtime.tls_g and the formula is different.

	var tls *elf.Prog
	for _, prog := range exe.Progs {
		if prog.Type == elf.PT_TLS {
			tls = prog
			break
		}
	}

	switch exe.Machine {
	case elf.EM_X86_64, elf.EM_386:
		tlsg := getSymbol(exe, "runtime.tlsg")
		if tlsg == nil || tls == nil {
			gStructOffset = ^uint64(PtrSize) + 1 //-ptrSize
			return
		}

		// According to https://reviews.llvm.org/D61824, linkers must pad the actual
		// size of the TLS segment to ensure that (tlsoffset%align) == (vaddr%align).
		// This formula, copied from the lld code, matches that.
		// https://github.com/llvm-mirror/lld/blob/9aef969544981d76bea8e4d1961d3a6980980ef9/ELF/InputSection.cpp#L643
		memsz := tls.Memsz + (-tls.Vaddr-tls.Memsz)&(tls.Align-1)

		// The TLS register points to the end of the TLS block, which is
		// tls.Memsz long. runtime.tlsg is an offset from the beginning of that block.
		gStructOffset = ^(memsz) + 1 + tlsg.Value // -tls.Memsz + tlsg.Value

	case elf.EM_AARCH64:
		tlsg := getSymbol(exe, "runtime.tls_g")
		if tlsg == nil || tls == nil {
			gStructOffset = 2 * uint64(PtrSize)
			return
		}

		gStructOffset = tlsg.Value + uint64(PtrSize*2) + ((tls.Vaddr - uint64(PtrSize*2)) & (tls.Align - 1))

	default:
		// we should never get here
		err = fmt.Errorf("architecture not supported")
	}

	return
}

func getGoidOffset(elfFile *elf.File) (goidOffset uint64, gStructOffset uint64, err error) {
	var dwarfData *dwarf.Data
	dwarfData, err = elfFile.DWARF()
	if err != nil {
		return
	}

	entryReader := dwarfData.Reader()

	var runtimeGOffset uint64
	var seenRuntimeG bool

	for {
		// Read all entries in sequence
		var entry *dwarf.Entry
		entry, err = entryReader.Next()
		if err == io.EOF || entry == nil {
			// We've reached the end of DWARF entries
			break
		}

		// Check if this entry is a struct
		if entry.Tag == dwarf.TagStructType {
			// Go through fields
			for _, field := range entry.Field {
				if field.Attr == dwarf.AttrName {
					val := field.Val.(string)
					if val == "runtime.g" {
						runtimeGOffset = uint64(entry.Offset)
						seenRuntimeG = true
					}
				}
			}
		}

		// Check if this entry is a struct member
		if seenRuntimeG && entry.Tag == dwarf.TagMember {
			// Go through fields
			for _, field := range entry.Field {
				if field.Attr == dwarf.AttrName {
					val := field.Val.(string)
					if val == "goid" {
						goidOffset = uint64(entry.Offset) - runtimeGOffset - 0x4b
						gStructOffset, err = getGStructOffset(elfFile)
						return
					}
				}
			}
		}
	}

	err = fmt.Errorf("goid not found in DWARF")
	return
}

func getOffsets(filePath string) (offsets map[string]*goExtendedOffset, goidOffset uint64, gStructOffset uint64, err error) {
	var engine gapstone.Engine
	switch runtime.GOARCH {
	case "amd64":
		engine, err = gapstone.New(
			gapstone.CS_ARCH_X86,
			gapstone.CS_MODE_64,
		)
	case "arm64":
		engine, err = gapstone.New(
			gapstone.CS_ARCH_ARM64,
			gapstone.CS_MODE_LITTLE_ENDIAN,
		)
	default:
		err = fmt.Errorf("Unsupported architecture: %v", runtime.GOARCH)
	}
	if err != nil {
		return
	}

	engineMajor, engineMinor := engine.Version()
	logger.Log.Infof(
		"Disassembling %s with Capstone %d.%d (arch: %d, mode: %d)",
		filePath,
		engineMajor,
		engineMinor,
		engine.Arch(),
		engine.Mode(),
	)

	offsets = make(map[string]*goExtendedOffset)
	var fd *os.File
	fd, err = os.Open(filePath)
	if err != nil {
		return
	}
	defer fd.Close()

	var elfFile *elf.File
	elfFile, err = elf.NewFile(fd)
	if err != nil {
		return
	}

	textSection := elfFile.Section(".text")
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

	var syms []elf.Symbol
	syms, err = elfFile.Symbols()
	if err != nil {
		return
	}
	for _, sym := range syms {
		offset := sym.Value

		var lastProg *elf.Prog
		for _, prog := range elfFile.Progs {
			if prog.Vaddr <= sym.Value && sym.Value < (prog.Vaddr+prog.Memsz) {
				offset = sym.Value - prog.Vaddr + prog.Off
				lastProg = prog
				break
			}
		}

		extendedOffset := &goExtendedOffset{enter: offset}

		// source: https://gist.github.com/grantseltzer/3efa8ecc5de1fb566e8091533050d608
		// skip over any symbols that aren't functions/methods
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
		textSectionDataLen := uint64(len(textSectionData) - 1)
		if symEndingIndex > textSectionDataLen {
			logger.Log.Warningf(
				"Skipping symbol %v, ending index %v is bigger than text section data length %v",
				sym.Name,
				symEndingIndex,
				textSectionDataLen,
			)
			continue
		}
		symBytes := textSectionData[symStartingIndex:symEndingIndex]

		// disassemble the symbol
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

	goidOffset, gStructOffset, err = getGoidOffset(elfFile)

	return
}

func getOffset(offsets map[string]*goExtendedOffset, symbol string) (*goExtendedOffset, error) {
	if offset, ok := offsets[symbol]; ok {
		return offset, nil
	}
	return nil, fmt.Errorf("symbol %s: %w", symbol, link.ErrNoSymbol)
}

func checkGoVersion(filePath string, offset *goExtendedOffset) (bool, string, error) {
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

	goVersionConstraint, err := semver.NewConstraint(fmt.Sprintf(">= %s", minimumABIInternalGoVersion))
	if err != nil {
		return false, goVersionStr, err
	}

	return goVersionConstraint.Check(goVersion), goVersionStr, nil
}
