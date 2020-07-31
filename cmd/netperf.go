package cmd

import (
	"../core"
)

var netperfTy string

func init() {
	rootCmd.PersistentFlags().StringVar(&netperfTy, "netperf-type", "tcp_rr", "netperf type benchmark")
	// TODO: add some validation here
}

// TODO: parse options to support other netperf configurations here
func getNetperfBench() core.Benchmark {
	switch netperfTy {
	case "tcp_rr":
		cnf := core.NetperfRRConf{core.NetperfConfDefault("tcp_rr")}
		cnf.Timeout = benchmarkDuration
		return &cnf

	case "np-rr":
		cnf := core.NetperfRRConf{core.NetperfConfDefault("tcp_rr")}
		cnf.Timeout = benchmarkDuration
		cnf.CliCommand = "scripts/np-rr.sh"
		return &cnf
	}

	panic("Fail!")
}
