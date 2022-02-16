#!/bin/bash

MIZU_HOME=$(realpath ../../../)

docker build -t mizu-ebpf-builder . || exit 1

docker run --rm \
	--name mizu-ebpf-builder \
	-v $MIZU_HOME:/mizu \
	-it mizu-ebpf-builder \
	sh -c "go generate tap/tlstapper/tls_tapper.go" || exit 1
