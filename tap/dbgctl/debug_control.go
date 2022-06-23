package dbgctl

import (
	"os"
)

var (
	MizuTapperDisablePcap              = os.Getenv("MIZU_DEBUG_DISABLE_PCAP") == "true"
	MizuTapperDisableTcpReassembly     = os.Getenv("MIZU_DEBUG_DISABLE_TCP_REASSEMBLY") == "true"
	MizuTapperDisableTcpStream         = os.Getenv("MIZU_DEBUG_DISABLE_TCP_STREAM") == "true"
	MizuTapperDisableDissectors        = os.Getenv("MIZU_DEBUG_DISABLE_DISSECTORS") == "true"
	MizuTapperDisableEmitting          = os.Getenv("MIZU_DEBUG_DISABLE_EMITTING") == "true"
	MizuTapperDisableSending           = os.Getenv("MIZU_DEBUG_DISABLE_SENDING") == "true"
	MizuTapperDisableNonHttpExtensions = os.Getenv("MIZU_DEBUG_DISABLE_NON_HTTP_EXTENSSION") == "true"
)
