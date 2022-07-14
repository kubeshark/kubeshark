#!/bin/bash

pushd "$(dirname "$0")" || exit 1

MIZU_HOME=$(realpath ../../../)

docker build -t mizu-ebpf-builder . || exit 1

BPF_TARGET=amd64
BPF_CFLAGS="-O2 -g -D__TARGET_ARCH_x86"
ARCH=$(uname -m)
if [[ $ARCH == "aarch64" || $ARCH == "arm64" ]]; then
    BPF_TARGET=arm64
    BPF_CFLAGS="-O2 -g -D__TARGET_ARCH_arm64"
fi

docker run --rm \
	--name mizu-ebpf-builder \
	-v $MIZU_HOME:/mizu \
	-v $(go env GOPATH):/root/go \
	mizu-ebpf-builder \
	sh -c "
		BPF_TARGET=\"$BPF_TARGET\" BPF_CFLAGS=\"$BPF_CFLAGS\" go generate tap/tlstapper/tls_tapper.go
        BPF_TARGET=\"$BPF_TARGET\" go generate tap/xdp/ipproto.go
        chown $(id -u):$(id -g) tap/tlstapper/tlstapper*_bpf*
	" || exit 1

popd
