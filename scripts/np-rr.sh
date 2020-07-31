#!/bin/bash
# wrapper script for netperf, intended to be used for RR benchmarks

# queue depth: packets in flight
QD=${QD:-"1 2 4 8 16 32 64 128"}

# packet (req/res) sizes
RR_SIZES=${RR_SIZES:-"1,1"}

for qd in ${QD}; do
    for rr_size in ${RR_SIZES}; do
        echo netperf "$@" -r $rr_size -b $qd
             netperf "$@" -r $rr_size -b $qd
    done
done
