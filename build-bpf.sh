#!/bin/bash

BPF_TARGET=amd64
BPF_CFLAGS="-O2 -g -D__TARGET_ARCH_x86"
ARCH=$(uname -m)
if [[ $ARCH == "aarch64" || $ARCH == "arm64" ]]; then
    BPF_TARGET=arm64
    BPF_CFLAGS="-O2 -g -D__TARGET_ARCH_arm64"
fi

BPF_TARGET=\"$BPF_TARGET\" BPF_CFLAGS=\"$BPF_CFLAGS\" go generate tap/tlstapper/tls_tapper.go
