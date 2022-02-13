#!/bin/bash

MIZU_BRANCH=$1
[ -z $MIZU_BRANCH ] && MIZU_BRANCH="develop"

docker build --build-arg "MIZU_BRANCH=$MIZU_BRANCH" -t mizu-ebpf-builder . || exit 1

mkdir -p output

docker run --rm \
	--name mizu-ebpf-builder \
	-v /usr/include/:/usr/include \
	-v $(pwd)/output:/output \
	-it mizu-ebpf-builder \
	bash -c "
		[ ! -r tap/tlstapper/tls_tapper.go ] && { echo File not found tap/tlstapper/tls_tapper.go; exit 1; }
		go generate tap/tlstapper/tls_tapper.go
		cp tap/tlstapper/tlstapper_bpfeb.o /output
		cp tap/tlstapper/tlstapper_bpfel.o /output
	" || exit 1

cp output/*.o ../
