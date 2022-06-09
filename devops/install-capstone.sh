#!/bin/bash

SUDO=''
if (( $EUID != 0 )); then
    SUDO='sudo'
fi

curl https://github.com/aquynh/capstone/archive/4.0.2.tar.gz -Lo ./capstone.tar.gz \
 && tar -xzf capstone.tar.gz && mv ./capstone-* ./capstone \
 && cd capstone \
 && ./make.sh \
 && $SUDO ./make.sh install
