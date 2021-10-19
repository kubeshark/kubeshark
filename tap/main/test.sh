#!/bin/bash

set -e

go build -o main main/main.go

sudo ./main/main "$@"
