#!/bin/bash

[ "$(ls -A --ignore='.??*' bin)" ] && echo "Skipping downloading BINs" || gsutil -m cp gs://static.up9.io/mizu/test-pcap/bin/http/\*.bin bin

go test *.go -v -covermode=atomic -coverprofile=coverage.out

TEST_UPDATE=1 go test *.go -v -covermode=atomic -coverprofile=coverage.out
