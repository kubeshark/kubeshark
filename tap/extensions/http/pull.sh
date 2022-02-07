#!/bin/bash

[ "$(ls -A --ignore='.??*' bin)" ] && echo "Skipping downloading BINs" || gsutil -m cp gs://static.up9.io/mizu/test-pcap/bin/http/\*.bin bin
