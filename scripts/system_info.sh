#!/bin/sh

set -x
set -o pipefail

uname -a

cat /host/boot/config-$(uname -r)
cat /host/etc/lsb-release

cat /proc/cpuinfo

(ip -j link  2>/dev/null | jq) || ip link
(ip -j addr  2>/dev/null | jq) || ip addr
(ip -j route 2>/dev/null | jq) || ip route

# ignore errors
exit 0
