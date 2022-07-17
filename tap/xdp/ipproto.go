package ebpf

import (
	"github.com/asavie/xdp"
	"github.com/cilium/ebpf"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go@v0.9.0 -target $BPF_TARGET ipproto ipproto.c -- -I/usr/include/ -I./include -nostdinc -O3

// NewIPProtoProgram returns an new eBPF that directs packets of the given ip protocol to to XDP sockets
func NewIPProtoProgram(protocol uint8, options *ebpf.CollectionOptions) (*xdp.Program, error) {
	spec, err := loadIpproto()
	if err != nil {
		return nil, err
	}

	if err := spec.RewriteConstants(map[string]interface{}{"PROTO": protocol}); err != nil {
		return nil, err
	}
	var program ipprotoObjects
	if err := spec.LoadAndAssign(&program, options); err != nil {
		return nil, err
	}

	p := &xdp.Program{Program: program.XdpSockProg, Queues: program.QidconfMap, Sockets: program.XsksMap}
	return p, nil
}
