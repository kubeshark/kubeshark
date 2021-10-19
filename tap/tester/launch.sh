#!/bin/bash

set -e

echo "Building extensions..."
pushd .. && ./devops/build_extensions.sh && popd

go build -o tester tester/tester.go

sudo ./tester/tester "$@"
