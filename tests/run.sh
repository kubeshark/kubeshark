#!/bin/bash

rm -rf entries/ && mkdir -p entries && \
rm -rf pprof/* && make clean && make agent || exit 1
setsid sh -c 'basenine -port 9099 & ./agent/build/mizuagent --api-server' &
PGID_MAIN=$! && \
# Wait for Basenine to be up
timeout 3 sh -c 'until nc -z $0 $1; do sleep 0.1; done' localhost 9099 && \
# Wait for Mizu API Server to be up
timeout 5 sh -c 'until nc -z $0 $1; do sleep 0.1; done' localhost 8899

PCAPS="tests/pcaps/*"
for file in $PCAPS
do
    echo "Dissecting $file"
    pcap=$file setsid sh -c 'MIZU_TEST=1 GOGC=12800 NODE_NAME=dev ./agent/build/mizuagent -r $pcap --tap --api-server-address ws://localhost:8899/wsTapper' && \
    sleep 1
done

sleep 5 && \

kill -TERM -- -$PGID_MAIN
