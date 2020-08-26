
Kubenetbench is a utility for benchmarking kubernetes networking.

Still under heavy development.

# Build


```
make
```

# Usage

kubenetbench works by executing `kubectl` commands, so it depends on this
working properly.

## Start a session

First, initalize a session
```
./kubenetbench/kubenetbench -s test init
2020/08/26 16:45:21 ================> wrote wrapper script: you may use: ./test/knb
2020/08/26 16:45:21 ****** ./kubenetbench/kubenetbench -s test init
2020/08/26 16:45:21 Starting session monitor
2020/08/26 16:45:21 Generating ./test/monitor.yaml
2020/08/26 16:45:21 $ kubectl apply -f ./test/monitor.yaml
2020/08/26 16:45:21 calling GetSysInfoNode on k8s1 (remaining retries: 10)
2020/08/26 16:45:25 calling GetSysInfoNode on k8s1 (remaining retries: 9)
2020/08/26 16:45:25 calling GetSysInfoNode on k8s2 (remaining retries: 10)
```

This will create a `./test` directory and spawn a monitor on all nodes of the
cluster as a daemonset (see: `test/monitor.yaml`). The monitor runs in
privileged mode and is used to collect system information and  potentially
prepare the nodes (absolutely no care was taken to make it safe, so be advised).

As a simple example, for each nodea system info file is created before any
benchmarking happens:

```
$ cat test/*.sysinfo
+ uname -a
Linux k8s1 5.8.0-rc1+ #1 SMP Wed Jun 24 08:02:36 UTC 2020 x86_64 Linux
+ uname -a
Linux k8s2 5.8.0-rc1+ #1 SMP Wed Jun 24 08:02:36 UTC 2020 x86_64 Linux
```

## Execute a benchmark

For convinience, a wrapper script (`test/knb`) is placed in the session
directory that users are expected to use.

A benchmark run consists of:
 * a k8s setup: currently two exist: `pod2pod` and `service`
 * the underlying benchmark, currently only `netperf` is supported

To run a pod-to-pod benchmark:

```
callisto~/go/src/github.com/kkourt/kubenetbench> ./test/knb pod2pod
2020/08/26 16:58:47 ****** /home/kkourt/go/src/github.com/kkourt/kubenetbench/kubenetbench/kubenetbench --session-id test --session-base-dir . pod2pod
2020/08/26 16:58:47 Generating ./test/pod2pod-20200826165847/netserv.yaml
2020/08/26 16:58:47 $ kubectl apply -f ./test/pod2pod-20200826165847/netserv.yaml
2020/08/26 16:58:49 $ kubectl get pod -l "knb-runid=pod2pod-20200826165847,role=srv" -o custom-columns=IP:.status.podIP --no-headers # (remaining retries: 10)
2020/08/26 16:58:51 $ kubectl get pod -l "knb-runid=pod2pod-20200826165847,role=srv" -o custom-columns=IP:.status.podIP --no-headers # (remaining retries: 9)
2020/08/26 16:58:51 server_ip=10.17.178.131 2020/08/26 16:58:51 Generating ./test/pod2pod-20200826165847/client.yaml
2020/08/26 16:58:51 $ kubectl apply -f ./test/pod2pod-20200826165847/client.yaml
2020/08/26 16:59:27 client phase: Succeeded 2020/08/26 16:59:27 $ kubectl logs knb-cli >
./test/pod2pod-20200826165847/cli.log 2020/08/26 16:59:27 $ kubectl logs knb-srv >
./test/pod2pod-20200826165847/srv.log 2020/08/26 16:59:27 $ kubectl delete pod,deployment,service,networkpolicy -l "knb-runid=pod2pod-20200826165847"
$ cat test/pod2pod-20200826165847/cli.log
MIGRATED TCP REQUEST/RESPONSE TEST from 0.0.0.0 (0.0.0.0) port 0 AF_INET to
10.17.178.131 () port 8000 AF_INET : demo : first burst 0
enable_enobufs failed: getprotobyname
THROUGHPUT=2841.07
THROUGHPUT_UNITS=Trans/s
TRANSACTION_RATE=2841.067
P50_LATENCY=334
P90_LATENCY=391
RT_LATENCY=351.980
MEAN_LATENCY=351.66
REQUEST_SIZE=1
RESPONSE_SIZE=1
LOCAL_TRANSPORT_RETRANS=0
REMOTE_TRANSPORT_RETRANS=0
```

Results in this case are placed in the `test/pod2pod-20200826165847` folder. The
above will run a `tcp_rr` netperf benchmark by default.

It is also possible to pass arbitrary arguments to the netperf benchmark using
`--netperf-args` and `--netperf-bench-args`. For example:
```
./kubenetbench pod2pod --runid foo --benchmark netperf --netperf-args "-D" --netperf-args "10" --netperf-bench-args "-r" --netperf-bench-args "1,1" --netperf-bench-args "-b" --netperf-bench-args "10"
```

## node affinities

Users can specify affinities using the `--client-affinity` and/or
`--server-affinity` options.

## recording perf profiles

The monitor can be used to record perf profiles (using `perf record`) on the
nodes that the benchmark runs. For example:


```
$ test/knb pod2pod --collect-perf
2020/08/26 17:04:33 ****** /home/kkourt/go/src/github.com/kkourt/kubenetbench/kubenetbench/kubenetbench --session-id test --session-base-dir . pod2pod --collect-perf
2020/08/26 17:04:33 Generating ./test/pod2pod-20200826170433/netserv.yaml
2020/08/26 17:04:33 $ kubectl apply -f ./test/pod2pod-20200826170433/netserv.yaml
2020/08/26 17:04:36 $ kubectl get pod -l "knb-runid=pod2pod-20200826170433,role=srv" -o custom-columns=IP:.status.podIP --no-headers # (remaining retries: 10)
2020/08/26 17:04:38 $ kubectl get pod -l "knb-runid=pod2pod-20200826170433,role=srv" -o custom-columns=IP:.status.podIP --no-headers # (remaining retries: 9)
2020/08/26 17:04:38 server_ip=10.17.10.110
2020/08/26 17:04:38 Generating ./test/pod2pod-20200826170433/client.yaml
2020/08/26 17:04:38 $ kubectl apply -f ./test/pod2pod-20200826170433/client.yaml
2020/08/26 17:04:43 $ kubectl get pod -l "knb-runid=pod2pod-20200826170433" -o custom-columns=F0:.metadata.name,F1:.spec.nodeName,F2:.status.phase --no-headers
2020/08/26 17:04:43 Pods:
2020/08/26 17:04:43  [knb-cli k8s2 Running]
2020/08/26 17:04:43  [knb-srv k8s2 Running]
2020/08/26 17:04:43 started collection on monitor k8s2
2020/08/26 17:05:13 client phase: Succeeded
2020/08/26 17:05:50 perf data for k8s2 can be found in: ./test/pod2pod-20200826170433/perf-k8s2.tar.bz2
2020/08/26 17:05:50 $ kubectl logs knb-cli > ./test/pod2pod-20200826170433/cli.log
2020/08/26 17:05:51 $ kubectl logs knb-srv > ./test/pod2pod-20200826170433/srv.log
2020/08/26 17:05:51 $ kubectl delete pod,deployment,service,networkpolicy -l "knb-runid=pod2pod-20200826170433"
```

Note that in this case the pods where scheduled on the same node. The perf
tarball is created using `perf archive` so it also contains debugging symbols.

## Stopping the monitor

To stop the monitor, terminate the session:

```
$ ./test/knb done
2020/08/26 17:24:23 ****** /home/kkourt/go/src/github.com/kkourt/kubenetbench/kubenetbench/kubenetbench --session-id test --session-base-dir . done
2020/08/26 17:24:23 Starting session monitor
2020/08/26 17:24:23 $ kubectl delete daemonset -l "knb-sessid=test"
```


# Implementation notes

* kubenetbench talks to the monitor via GRPC
* monitor container is build using `Dockerfile.knb-monitor`
* netperf benchmark container is build using `Dockerfile.knb`
* YAML files are generated using templates and stored in the directory
