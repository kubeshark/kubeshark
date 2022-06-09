#!/bin/bash

SUDO=''
if (( $EUID != 0 )); then
    SUDO='sudo'
fi

git clone https://github.com/capstone-engine/capstone.git -b 4.0.2 --depth 1 && \
cd capstone && \
./make.sh && \
$SUDO ./make.sh install
