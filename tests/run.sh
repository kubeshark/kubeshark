#!/bin/bash

PCAPS_DIR=tests/pcaps
[ "$(ls -A --ignore='.??*' tests/pcaps)" ] && echo "Skipping downloading PCAPs" || gsutil cp gs://static.up9.io/mizu/test-pcap/\*.pcap $PCAPS_DIR

rm -rf entries/ && mkdir -p entries && \
rm -rf pprof/* && make clean && make agent || exit 1
setsid sh -c 'basenine -port 9099 & ./agent/build/mizuagent --api-server' &
PGID_MAIN=$! && \
# Wait for Basenine to be up
timeout 3 sh -c 'until nc -z $0 $1; do sleep 0.1; done' localhost 9099 && \
# Wait for Mizu API Server to be up
timeout 5 sh -c 'until nc -z $0 $1; do sleep 0.1; done' localhost 8899

PCAPS="$PCAPS_DIR/*"
for file in $PCAPS
do
    echo "Dissecting $file"
    pcap=$file timeout 5 sh -c 'MIZU_TEST=1 GOGC=12800 NODE_NAME=dev ./agent/build/mizuagent -r $pcap --tap --api-server-address ws://localhost:8899/wsTapper'
    exit_status=$?
    if [[ $exit_status -eq 124 ]]; then
        echo "Tapper timed out. Removing $file"
        rm $file
    fi
done

sleep 30 && \

kill -TERM -- -$PGID_MAIN
