#!/bin/bash
# wrapper script for netperf, intended to be used for RR benchmarks

# queue depth: packets in flight
QD=${QD:-"1 2 4 8 16 32 64 128"}

# packet (req/res) sizes
RR_SIZES=${RR_SIZES:-"1,1"}

# RR_TESTS="tcp_rr tcp_crr"

ARGS=()
TEST_ARGS=()

while [[ $# -gt 0 ]]; do
    opt=$1
    case $opt in
        -t) RR_TESTS=$(echo $2 | tr ',' ' ')
            shift 2
            ;;
        --)
            np_args_done=1
            shift 1
            ;;
        *)
            if [[ -n $np_args_done ]]; then
                TEST_ARGS+=("$1")
            else
                ARGS+=("$1")
            fi
            shift 1
            ;;
    esac
done


for rrt in ${RR_TESTS}; do
    for qd in ${QD}; do
        for rr_size in ${RR_SIZES}; do
            echo netperf ${ARGS[@]} -t $rrt -- ${TEST_ARGS[@]} -r $rr_size -b $qd
            if [ -z "$DRY_RUN" ]; then
                 netperf ${ARGS[@]} -t $rrt -- ${TEST_ARGS[@]} -r $rr_size -b $qd
            fi
            echo __DONE__
        done
    done
done
