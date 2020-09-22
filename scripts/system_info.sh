#!/bin/sh

set -x

uname -a

cat /boot/config-$(uname -r)
cat /etc/lsb-release

cat /proc/cpuinfo

ip -j link  | jq
ip -j addr  | jq
ip -j route | jq
