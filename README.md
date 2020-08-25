
Kubenetbench is a utility for benchmarking kubernetes networking.

**Text below is outdated**

# Build


```
go get -u github.com/spf13/cobra/cobra
go build
```

# Usage

* There are two main benchmark commands:
  * pod2pod: spawns a client and a server that communicate directly
  * service: the client accesses the server via a k8s service

* Each benchmark run has an identifier: `-i XXXXX`.

* Currently, there is only one benchmark supported: `-b netperf`. 
  * There are three types of netperf benchmarks: `--netperf-type`:
     * tcp_rr: https://hewlettpackard.github.io/netperf/doc/netperf.html#TCP_005fRR
     * tcp_crr: https://hewlettpackard.github.io/netperf/doc/netperf.html#TCP_005fCRR
     * script-np-rr: a script that performs multiple netperf runs
       * ([scripts/np-rr.sh](scripts/np-rr.sh))
       * NB: the container used has to be updated for changes in the script to take
         effect

# Example

```
$ ./kubenetbench pod2pod --runid foo --benchmark netperf
2020/08/10 11:23:32 Generating ./foo-20200810-112332/netserv.yaml
2020/08/10 11:23:32 $ kubectl apply -f ./foo-20200810-112332/netserv.yaml 
2020/08/10 11:23:35 $ kubectl get pod -l "kubenetbench-runid=foo,role=srv" -o custom-columns=IP:.status.podIP --no-headers # (remaining retries: 10)
2020/08/10 11:23:37 $ kubectl get pod -l "kubenetbench-runid=foo,role=srv" -o custom-columns=IP:.status.podIP --no-headers # (remaining retries: 9)
2020/08/10 11:23:39 $ kubectl get pod -l "kubenetbench-runid=foo,role=srv" -o custom-columns=IP:.status.podIP --no-headers # (remaining retries: 8)
2020/08/10 11:23:39 server_ip=10.17.73.153
2020/08/10 11:23:39 Generating ./foo-20200810-112332/client.yaml
2020/08/10 11:23:39 $ kubectl apply -f ./foo-20200810-112332/client.yaml 
2020/08/10 11:24:19 client phase: Succeeded
2020/08/10 11:24:19 $ kubectl logs kubenetbench-foo-cli > ./foo-20200810-112332/cli.log 
2020/08/10 11:24:20 $ kubectl logs kubenetbench-foo-srv > ./foo-20200810-112332/srv.log 
2020/08/10 11:24:20 $ kubectl delete pod,deployment,service,networkpolicy -l "kubenetbench-runid=foo"
```

The results of the run can be found in the  `./foo-20200810-112332` directory.
Specifically, `./foo-20200810-112332/cli.log` contains the output of the netperf
command.



It is also possible to pass arbitrary arguments to the netperf benchmark using
`--netperf-args` and `--netperf-bench-args`. For example:
```
./kubenetbench pod2pod --runid foo --benchmark netperf --netperf-args "-D" --netperf-args "10" --netperf-bench-args "-r" --netperf-bench-args "1,1" --netperf-bench-args "-b" --netperf-bench-args "10"
```
