#!/bin/bash
set -e

# Build it on arm64
docker build . -t kubeshark/linux-x86_64-musl-go-libpcap-capstone-bpf:capstone-5.0-rc2 && docker push kubeshark/linux-x86_64-musl-go-libpcap-capstone-bpf:capstone-5.0-rc2
