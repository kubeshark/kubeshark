#!/bin/sh

VER=test make cli && sudo cp ./cli/bin/kubeshark__ /usr/local/bin/kubeshark
