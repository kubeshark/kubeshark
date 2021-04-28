#!/bin/bash
./apiserver &
./passivetapper -i eth0 &
wait -n
pkill -P $$
