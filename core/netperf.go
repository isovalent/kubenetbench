package core

import (
	"fmt"

	"../utils"
)

type NetperfConf struct {
	Timeout  int
	DataPort uint16
	TestName string
}

type NetperfRRConf struct {
	NetperfConf
}

func (cnf *NetperfConf) GetTimeout() int {
	return cnf.Timeout
}

func (cnf *NetperfConf) WriteSrvYaml(pw *utils.PrefixWriter, params map[string]interface{}) {
	pw.AppendNewLineOrDie(`name: netperf-srv`)
	pw.AppendNewLineOrDie(`image: kkourt/netperf`)
	pw.AppendNewLineOrDie(`command: ["netserver"]`)
	pw.AppendNewLineOrDie(`args : [`)
	pw.PushPrefix("    ")
	pw.AppendNewLineOrDie(`"-D", # dont daemonize`)
	pw.PopPrefix()
	pw.AppendNewLineOrDie(`]`)
}

func (cnf *NetperfConf) WriteCliStart(pw *utils.PrefixWriter, params map[string]interface{}) {
	pw.AppendNewLineOrDie(`name: netperf-cli`)
	pw.AppendNewLineOrDie(`image: kkourt/netperf`)
	pw.AppendNewLineOrDie(`command: ["netperf"]`)
}

func (cnf *NetperfConf) WriteCliBaseArgs(pw *utils.PrefixWriter, params map[string]interface{}) {
	serverIP, ok := params["serverIP"]
	if !ok {
		panic("serverIP undefined")
	}

	pw.AppendNewLineOrDie(fmt.Sprintf(`"-l", "%d", # timeout`, cnf.Timeout))
	pw.AppendNewLineOrDie(`"-j", # enable additional statistics`)
	pw.AppendNewLineOrDie(fmt.Sprintf(`"-H", "%v",`, serverIP))
	pw.AppendNewLineOrDie(fmt.Sprintf(`"-t", "%s", # testname`, cnf.TestName))
}

func (cnf *NetperfRRConf) WriteCliYaml(pw *utils.PrefixWriter, params map[string]interface{}) {
	cnf.WriteCliStart(pw, params)
	pw.AppendNewLineOrDie(`args : [`)
	pw.PushPrefix("    ")
	cnf.WriteCliBaseArgs(pw, params)
	pw.AppendNewLineOrDie(`"--",`)
	pw.AppendNewLineOrDie(fmt.Sprintf(`"-P", "%d", # data connection port`, cnf.DataPort))
	pw.AppendNewLineOrDie(`"-k", "THROUGHPUT,THROUGHPUT_UNITS,P50_LATENCY,P99_LATENCY,REQUEST_SIZE,RESPONSE_SIZE",`)
	pw.PopPrefix()
	pw.AppendNewLineOrDie(`]`)
}
