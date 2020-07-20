#!/bin/bash

# Parameters
BENCH_TIMEOUT=10
DATA_PORT=8888

function usage() {
    echo "Usage: $0 [-i] run_id"
    echo "  Flags:"
    echo "    -i: isolated mode"
}

function generate_srv_yaml() {
# Generate server description
cat > $rundir/netserv.yaml <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: netperf-${runid}-srv
  labels : {
    runid: ${runid},
    role: srv,
  }
spec:
  containers:
  - name: netperf
    image: kkourt/netperf
    command: ["netserver"]
    # -D: dont daemonize
    args: ["-D"]
EOF
}

function generate_cli_yaml() {
    # arguments:
    srvip=$1

cat > $rundir/client.yaml <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: netperf-${runid}-cli
  labels : {
     runid: ${runid},
     role: cli,
  }
spec:
  restartPolicy: Never
  containers:
  - name: netperf
    image: kkourt/netperf
    command: ["netperf"]
    args: [
        "-l", "${BENCH_TIMEOUT}", # timeout
        "-j",                     # enable additional statistics
        "-H", "${srvip}",         # server IP
        "-t", "tcp_rr",           # test name
        "--",                     # test specific arguemnts
        "-P", "${DATA_PORT}",     # data connection port
        # additional metrics to record
        "-k", "THROUGHPUT,THROUGHPUT_UNITS,P50_LATENCY,P99_LATENCY,REQUEST_SIZE,RESPONSE_SIZE"
    ]
EOF
}

function generate_policy_yaml() {
cat > $rundir/netpolicy.yaml << EOF
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: netperf-${runid}-policy
  labels : {
     "runid": ${runid},
  }
spec:
  podSelector:
    matchLabels:
      role: srv
  policyTypes:
  - Ingress
  ingress:
  - from:
    ports:
    # control port
    - protocol: TCP
      port: 12865
    # data port
    - protocol: TCP
      port: $DATA_PORT
EOF
}

function get_pod_ip_selector() {
    # arguments:
    selector=$1

    nretries=0
    while true; do
        ret=$(kubectl get -l "$selector" pod -o custom-columns=IP:.status.podIP --no-headers)
        if [ $ret != "<none>" ]; then
            break
        fi
        let nretries=nretries+1
        if [ $nretries -eq 10 ]; then
            echo "Maxium number of retries to get server IP reached. Bailing out." 1>&2
            exit 1
        fi
        # echo "Retrying ..."
        sleep 1s
    done

    echo $ret
}

while getopts "i" option
do
        case $option in
                i) ISOLATED=1 ;;
                *) echo "Unknown option"; usage; exit 1 ;;
        esac
done

shift $((OPTIND-1))
if [ -z "$1" ]; then
    usage
    exit 1
fi

set -x

runid=$1
rundir=${runid}-$(date +%Y%m%d.%H%M%S)

mkdir $rundir

# run server
generate_srv_yaml
kubectl apply -f $rundir/netserv.yaml

# get server IP
sleep 2s
srvip=$(get_pod_ip_selector "runid=${runid},role=srv")

if [ -n "$ISOLATED" ]; then
        generate_policy_yaml
fi


# run client
generate_cli_yaml $srvip
kubectl apply -f $rundir/client.yaml

# wait until client is done
sleep $(( $BENCH_TIMEOUT + 5))s
nretries=0
while true; do
    cli_phase=$(kubectl get -l "runid=${runid},role=cli" pod -o custom-columns=Status:.status.phase --no-headers)
    if [ $cli_phase == "Succeeded" ] || [ $cli_phase == "Failed" ]; then
            break
    fi

    let nretries=nretries+1
    if [ $nretries -eq 10 ]; then
        echo "Maxium number of retries to wait for client reached. Bailing out."
        break
    fi
    # echo "Retrying ..."
    sleep 1s
done

# get logs
kubectl logs netperf-${runid}-srv  > ${rundir}/srv.log
kubectl logs netperf-${runid}-cli  > ${rundir}/cli.log
kubectl delete pod -l "runid=${runid}"
kubectl delete networkpolicy -l "runid=${runid}"
