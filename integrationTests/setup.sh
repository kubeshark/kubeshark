#!/bin/bash

# TODO: ARCH parameter to use on basenine and build agent

PREFIX=/usr/local/bin
VERSION=v0.4.13
GOARCH=amd64

echo "Attempting to install basenine $GOARCH to $PREFIX"

# Download Basenine executable, verify the sha1sum
curl -LO "https://github.com/up9inc/basenine/releases/download/$VERSION/basenine_linux_$GOARCH"
curl -LO "https://github.com/up9inc/basenine/releases/download/$VERSION/basenine_linux_$GOARCH.sha256"
shasum -a 256 -c basenine_linux_$GOARCH.sha256
rm -rf basenine_linux_$GOARCH.sha256
chmod +x basenine_linux_$GOARCH
mv basenine_linux_$GOARCH basenine
mv basenine "$PREFIX"

echo "Build agent"
sudo apt-get install libpcap-dev
rm -rf entries/ && mkdir -p entries && rm -rf pprof/*
make agent