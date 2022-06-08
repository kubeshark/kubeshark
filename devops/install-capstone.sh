#!/bin/bash

git clone https://github.com/capstone-engine/capstone.git -b 4.0.2 && \
git checkout capstone && \
./make.sh && \
sudo ./make.sh install
