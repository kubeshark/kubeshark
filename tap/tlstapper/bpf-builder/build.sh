#!/bin/bash

pushd "$(dirname "$0")" || exit 1

MIZU_HOME=$(realpath ../../../)

docker build -t mizu-ebpf-builder . || exit 1

BPF_TARGET=amd64
BPF_CFLAGS="-O2 -g -D__TARGET_ARCH_x86"
ARCH=$(uname -m)
if [[ $ARCH == "aarch64" ]]; then
    BPF_TARGET=arm64
    BPF_CFLAGS="-O2 -g -D__TARGET_ARCH_arm64"
fi

docker run --rm \
	--name mizu-ebpf-builder \
	-v $MIZU_HOME:/mizu \
	-v $(go env GOPATH):/root/go \
	-it mizu-ebpf-builder \
	sh -c "
		BPF_TARGET=\"$BPF_TARGET\" BPF_CFLAGS=\"$BPF_CFLAGS\" go generate tap/tlstapper/tls_tapper.go
		chown $(id -u):$(id -g) tap/tlstapper/tlstapper_bpf*
	" || exit 1

popd
