#!/bin/bash
set -e

# Build it on x86_64
docker build . -t up9inc/linux-arm64-musl-go-libpcap-capstone-bpf:capstone-5.0-rc2 && docker push up9inc/linux-arm64-musl-go-libpcap-capstone-bpf:capstone-5.0-rc2
