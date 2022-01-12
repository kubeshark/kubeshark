#!/bin/bash

for f in tap/extensions/*; do
    if [ -d "$f" ]; then
        echo Building extension: $f
        extension=$(basename $f) && \
        cd tap/extensions/${extension} && \
        go build -buildmode=plugin -o ../${extension}.so .  && \
        cd ../../..  && \
        mkdir -p agent/build/extensions  && \
        cp tap/extensions/${extension}.so agent/build/extensions
    fi
done
