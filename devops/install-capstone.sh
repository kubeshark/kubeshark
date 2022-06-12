#!/bin/bash

SUDO=''
if (( $EUID != 0 )); then
    SUDO='sudo'
fi

curl https://github.com/capstone-engine/capstone/archive/4.0.2.tar.gz -Lo ./capstone.tar.gz \
 && tar -xzf capstone.tar.gz && mv ./capstone-* ./capstone \
 && cd capstone \
 && CAPSTONE_ARCHS="aarch64 x86" ./make.sh \
 && $SUDO ./make.sh install
