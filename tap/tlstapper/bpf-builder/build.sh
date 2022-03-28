#!/bin/bash

MIZU_HOME=$(realpath ../../../)

docker build -t mizu-ebpf-builder . || exit 1

docker run --rm \
	--name mizu-ebpf-builder \
	-v $MIZU_HOME:/mizu \
	-it mizu-ebpf-builder \
	sh -c "
		go generate tap/tlstapper/tls_tapper.go
		chown $(id -u):$(id -g) tap/tlstapper/tlstapper_bpfeb.go
		chown $(id -u):$(id -g) tap/tlstapper/tlstapper_bpfeb.o
		chown $(id -u):$(id -g) tap/tlstapper/tlstapper_bpfel.go
		chown $(id -u):$(id -g) tap/tlstapper/tlstapper_bpfel.o
	" || exit 1
