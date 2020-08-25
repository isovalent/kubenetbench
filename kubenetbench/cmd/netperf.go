package cmd

import (
	"github.com/spf13/cobra"

	"github.com/kkourt/kubenetbench/kubenetbench/core"
)

var netperfTy string
var netperfArgs []string
var netperfBenchArgs []string

func init() {
}

func addNetperfFlags(cmd *cobra.Command) {
	// TODO: add some validation here
	cmd.Flags().StringVar(&netperfTy, "netperf-type", "tcp_rr", "netperf type benchmark (tcp_rr, tcp_crr, script-np-rr)")
	cmd.Flags().StringArrayVar(&netperfArgs, "netperf-args", []string{}, "netperf arguments")
	cmd.Flags().StringArrayVar(&netperfBenchArgs, "netperf-bench-args", []string{}, "netperf benchmark arguments (after --)")
}

// TODO: parse options to support other netperf configurations here
func getNetperfBench() core.Benchmark {
	switch netperfTy {
	case "tcp_rr":
		cnf := core.NetperfRRConf{core.NetperfConfDefault("tcp_rr", netperfArgs, netperfBenchArgs)}
		cnf.Timeout = benchmarkDuration
		return &cnf

	case "tcp_crr":
		cnf := core.NetperfRRConf{core.NetperfConfDefault("tcp_crr", netperfArgs, netperfBenchArgs)}
		cnf.Timeout = benchmarkDuration
		return &cnf

	case "script-np-rr":
		cnf := core.NetperfRRConf{core.NetperfConfDefault("tcp_rr,udp_rr,tcp_crr", netperfArgs, netperfBenchArgs)}
		cnf.Timeout = benchmarkDuration
		cnf.CliCommand = "scripts/np-rr.sh"
		return &cnf
	}

	panic("Fail!")
}
