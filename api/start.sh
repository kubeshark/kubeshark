#!/bin/bash
./mizuagent -i any -hardump -hardir /tmp/mizuhars -harentriesperfile 1 -targets ${TAPPED_ADDRESSES}
## this script runs both executables and exits everything if one fails
#./apiserver -hardir /tmp/mizuhars &
#./passivetapper -i any -hardump -hardir /tmp/mizuhars -harentriesperfile 5 -targets "${TAPPED_ADDRESSES}" &
#wait -n
# pkill -P $$
