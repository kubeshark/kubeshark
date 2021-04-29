#!/bin/bash

# this script runs both executables and exits everything if one fails
./apiserver -hardir /tmp/mizuhars &
./passivetapper -i any -hardump -hardir /tmp/mizuhars -harentriesperfile 5 -targets "${TAPPED_ADDRESSES}" &
wait -n
pkill -P $$
