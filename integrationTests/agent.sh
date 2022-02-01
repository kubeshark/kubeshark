#!/bin/bash

if [ "$EUID" -eq 0 ]
  then echo "Run without sudo!"
  exit
fi

if [ "$1" = "" ]
  then echo "Missing config file path!"
  exit
fi

if [[ $1 != *.json ]]
  then echo "Config file must be a .json file!"
  exit
fi

CONFIG=$PWD/$1
echo "Config file path: $CONFIG"

if [ "$2" = "" ]
  then echo "Missing pcap file path!"
  exit
fi

if [[ $2 != *.cap ]]
  then echo "PCAP file must be a .cap file!"
  exit
fi

PCAP=$PWD/$2
echo "PCAP file path: $PCAP"

rm -rf entries/ && mkdir -p entries && rm -rf pprof/* && \
make agent && \
basenine -port 9099 & \
PID1=$! && \
./agent/build/mizuagent \
    --config-path $CONFIG \
    --api-server & \
PID2=$! && \
sleep 0.5 && \
GOGC=12800 NODE_NAME=dev \
    ./agent/build/mizuagent \
        -r $PCAP \
        --tap \
        --api-server-address ws://localhost:8899/wsTapper & \
PID3=$! && \
read -r -d '' _ </dev/tty
kill -9 $PID1 && \
kill -9 $PID2 && \
kill -9 $PID3