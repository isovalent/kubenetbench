#!/bin/sh

xid=$1

if [ -z $xid ]; then
    echo "Usage: $0 <xid>"
    exit 1
fi

$(dirname $0)/perf-archive.sh /tmp/$xid-perf.data
