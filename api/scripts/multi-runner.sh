#!/bin/bash

# this script runs both executables and exits everything if one fails
./apiserver &
./passivetapper -i any -hardump -hardir /tmp/mizuhars -harentriesperfile 50 &
wait -n
pkill -P $$
