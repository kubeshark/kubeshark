#!/bin/bash
set -e

# Build it on x86_64
docker build . -t kubeshark/linux-arm64-musl-go-libpcap-capstone-bpf:capstone-5.0-rc2 && docker push kubeshark/linux-arm64-musl-go-libpcap-capstone-bpf:capstone-5.0-rc2
