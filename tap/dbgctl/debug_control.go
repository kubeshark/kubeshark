package dbgctl

import (
	"os"
)

var (
	KubesharkTapperDisablePcap              = os.Getenv("KUBESHARK_DEBUG_DISABLE_PCAP") == "true"
	KubesharkTapperDisableTcpReassembly     = os.Getenv("KUBESHARK_DEBUG_DISABLE_TCP_REASSEMBLY") == "true"
	KubesharkTapperDisableTcpStream         = os.Getenv("KUBESHARK_DEBUG_DISABLE_TCP_STREAM") == "true"
	KubesharkTapperDisableDissectors        = os.Getenv("KUBESHARK_DEBUG_DISABLE_DISSECTORS") == "true"
	KubesharkTapperDisableEmitting          = os.Getenv("KUBESHARK_DEBUG_DISABLE_EMITTING") == "true"
	KubesharkTapperDisableSending           = os.Getenv("KUBESHARK_DEBUG_DISABLE_SENDING") == "true"
	KubesharkTapperDisableNonHttpExtensions = os.Getenv("KUBESHARK_DEBUG_DISABLE_NON_HTTP_EXTENSSION") == "true"
)
