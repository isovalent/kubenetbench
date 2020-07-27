package cmd

import (
	"../core"
)

// NB: eventually we will probably want to parse netperf specific arguments here

func getNetperfConf() *core.NetperfConf {
	return core.NetperfConfDefault()
}
