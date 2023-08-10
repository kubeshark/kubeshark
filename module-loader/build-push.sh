#!/bin/bash
set -e

# Build it on x86_64
docker build . -t kubeshark/module-loader:latest --build-arg KERNEL_VERSION="5.10.0-23-amd64" && docker push kubeshark/module-loader:latest
