package cmd

import (
	"../core"
)

var netperfTy string

func init() {
	rootCmd.PersistentFlags().StringVar(&netperfTy, "netperf-type", "tcp_rr", "netperf type benchmark (tcp_rr, tcp_crr, script-np-rr)")
	// TODO: add some validation here
}

// TODO: parse options to support other netperf configurations here
func getNetperfBench() core.Benchmark {
	switch netperfTy {
	case "tcp_rr":
		cnf := core.NetperfRRConf{core.NetperfConfDefault("tcp_rr")}
		cnf.Timeout = benchmarkDuration
		return &cnf

	case "tcp_crr":
		cnf := core.NetperfRRConf{core.NetperfConfDefault("tcp_crr")}
		cnf.Timeout = benchmarkDuration
		return &cnf

	case "script-np-rr":
		cnf := core.NetperfRRConf{core.NetperfConfDefault("tcp_rr,udp_rr,tcp_crr")}
		cnf.Timeout = benchmarkDuration
		cnf.CliCommand = "scripts/np-rr.sh"
		return &cnf
	}

	panic("Fail!")
}
