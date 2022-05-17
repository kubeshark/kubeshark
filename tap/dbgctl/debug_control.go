package dbgctl

import (
	"os"
)

var (
	MizuTapperDisablePcap          bool = os.Getenv("MIZU_TAPPER_DISABLE_PCAP") == "true"
	MizuTapperDisableTcpReassembly bool = os.Getenv("MIZU_TAPPER_DISABLE_TCP_REASSEMBLY") == "true"
	MizuTapperDisableTapTarget     bool = os.Getenv("MIZU_TAPPER_DISABLE_TAP_TARGET") == "true"
	MizuTapperDisableDissectors    bool = os.Getenv("MIZU_TAPPER_DISABLE_DISSECTORS") == "true"
	MizuTapperDisableEmitting      bool = os.Getenv("MIZU_TAPPER_DISABLE_EMITTING") == "true"
	MizuTapperDisableSending       bool = os.Getenv("MIZU_TAPPER_DISABLE_SENDING") == "true"
)
