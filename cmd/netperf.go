package cmd

import (
	"../core"
)

// TODO: parse options to support other netperf configurations here
func getNetperfBench() core.Benchmark {
	return &core.NetperfRRConf{
		core.NetperfConf{
			Timeout:  10,
			DataPort: 8000,
			TestName: "tcp_rr",
		},
	}
}
