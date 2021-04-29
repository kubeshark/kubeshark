#!/bin/bash

# this script runs both executables and exits everything if one fails
./apiserver &
./passivetapper -i eth0 &
wait -n
pkill -P $$
