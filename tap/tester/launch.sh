#!/bin/bash

set -e

go build -o tester tester/tester.go

sudo ./tester/tester "$@"
