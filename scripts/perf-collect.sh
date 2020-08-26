#!/bin/sh

xid=$1


if [ -z $xid ]; then
    echo "Usage: $0 <xid>"
    exit 1
fi

xdir=$(mktemp -d /tmp/perf-collect.XXXXXXX)
echo $xdir
cd $xdir
cp /tmp/$xid-perf.data .
$(dirname $0)/perf-archive.sh $xid-perf.data >log.out 2>log.err

tar cjf /tmp/$xid-perf.data.tar.bz2 .
