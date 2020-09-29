package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cilium/kubenetbench/kubenetbench/core"
)

var netperfTy string
var netperfArgs []string
var netperfBenchArgs []string
var netperfNStreams int

var netperfBenchMap = map[string]func() core.Benchmark{
	"tcp_rr": func() core.Benchmark {
		cnf := core.NetperfRRConf{core.NetperfConfDefault("tcp_rr", netperfArgs, netperfBenchArgs)}
		cnf.Timeout = benchmarkDuration
		handle_nstreams(&cnf.NetperfConf)
		return &cnf
	},

	"tcp_crr": func() core.Benchmark {
		cnf := core.NetperfRRConf{core.NetperfConfDefault("tcp_crr", netperfArgs, netperfBenchArgs)}
		cnf.Timeout = benchmarkDuration
		handle_nstreams(&cnf.NetperfConf)
		return &cnf
	},

	"udp_rr": func() core.Benchmark {
		cnf := core.NetperfRRConf{core.NetperfConfDefault("udp_rr", netperfArgs, netperfBenchArgs)}
		cnf.Timeout = benchmarkDuration
		handle_nstreams(&cnf.NetperfConf)
		return &cnf
	},

	"tcp_stream": func() core.Benchmark {
		cnf := core.NetperfStreamConf{core.NetperfConfDefault("tcp_stream", netperfArgs, netperfBenchArgs)}
		cnf.Timeout = benchmarkDuration
		handle_nstreams(&cnf.NetperfConf)
		return &cnf
	},

	"tcp_maerts": func() core.Benchmark {
		cnf := core.NetperfStreamConf{core.NetperfConfDefault("tcp_maerts", netperfArgs, netperfBenchArgs)}
		cnf.Timeout = benchmarkDuration
		handle_nstreams(&cnf.NetperfConf)
		return &cnf
	},

	"udp_stream": func() core.Benchmark {
		cnf := core.NetperfStreamConf{core.NetperfConfDefault("udp_stream", netperfArgs, netperfBenchArgs)}
		cnf.Timeout = benchmarkDuration
		handle_nstreams(&cnf.NetperfConf)
		return &cnf
	},
}

func addNetperfFlags(cmd *cobra.Command) {
	// TODO: add some validation here
	tys := make([]string, 0, len(netperfBenchMap))
	for ty := range netperfBenchMap {
		tys = append(tys, ty)
	}

	tyHelp := fmt.Sprintf("netperf type benchmark (available values: %s)", strings.Join(tys, ","))
	cmd.Flags().StringVar(&netperfTy, "netperf-type", "tcp_rr", tyHelp)
	cmd.Flags().StringArrayVar(&netperfArgs, "netperf-args", []string{}, "netperf arguments")
	cmd.Flags().StringArrayVar(&netperfBenchArgs, "netperf-bench-args", []string{}, "netperf benchmark arguments (after --)")
	cmd.Flags().IntVar(&netperfNStreams, "netperf-nstreams", 0, ">0 value enables using duper_netperf script for multiple streams")
}

func handle_nstreams(conf *core.NetperfConf) {
	if netperfNStreams == 0 {
		return
	}

	if conf.CliCommand == "netperf" {
		conf.CliCommand = "scripts/duper_netperf"
		conf.PreArgs = append(conf.PreArgs, fmt.Sprintf("%d", netperfNStreams))
	} else {
		log.Fatal("cannot use multiple streams with CliCommand=%s", conf.CliCommand)
	}
}

// TODO: parse options to support other netperf configurations here
func getNetperfBench() core.Benchmark {

	initFn, ok := netperfBenchMap[netperfTy]
	if !ok {
		panic(fmt.Sprintf("Invalid netperf type: %s", netperfTy))
	}
	return initFn()
}
