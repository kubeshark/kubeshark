package dbgctl

import (
	"os"
)

var (
	MizuTapperDisablePcap              bool = os.Getenv("MIZU_DEBUG_DISABLE_PCAP") == "true"
	MizuTapperDisableTcpReassembly     bool = os.Getenv("MIZU_DEBUG_DISABLE_TCP_REASSEMBLY") == "true"
	MizuTapperDisableTcpStream         bool = os.Getenv("MIZU_DEBUG_DISABLE_TCP_STREAM") == "true"
	MizuTapperDisableDissectors        bool = os.Getenv("MIZU_DEBUG_DISABLE_DISSECTORS") == "true"
	MizuTapperDisableEmitting          bool = os.Getenv("MIZU_DEBUG_DISABLE_EMITTING") == "true"
	MizuTapperDisableSending           bool = os.Getenv("MIZU_DEBUG_DISABLE_SENDING") == "true"
	MizuTapperDisableNonHttpExtensions bool = os.Getenv("MIZU_DEBUG_DISABLE_NON_HTTP_EXTENSSION") == "true"
)
