#!/bin/sh

timeout=$1
xid=$2

if [ -z $xid ]; then
    echo "Usage: $0 <timeout> <xid>"
    exit 1
fi

set -x
perf record -g -a -o /tmp/$xid-perf.data sleep $timeout
