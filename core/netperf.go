package core

import (
	"fmt"

	"../utils"
)

// NetperfConf base netperf configuration
type NetperfConf struct {
	Timeout  int
	DataPort uint16
	TestName string
}

// NetperfRRConf RR netperf configuration
type NetperfRRConf struct {
	NetperfConf
}

// GetTimeout returns the benchmark timeout
func (cnf *NetperfConf) GetTimeout() int {
	return cnf.Timeout
}

// WriteSrvContainerYaml writes the server yaml
func (cnf *NetperfConf) WriteSrvContainerYaml(pw *utils.PrefixWriter, params map[string]interface{}) {
	pw.AppendNewLineOrDie(`name: netperf-srv`)
	pw.AppendNewLineOrDie(`image: kkourt/netperf`)
	pw.AppendNewLineOrDie(`command: ["netserver"]`)
	pw.AppendNewLineOrDie(`args : [`)
	pw.PushPrefix("    ")
	pw.AppendNewLineOrDie(`"-D", # dont daemonize`)
	pw.PopPrefix()
	pw.AppendNewLineOrDie(`]`)
}

func (cnf *NetperfConf) writeCliStart(pw *utils.PrefixWriter, params map[string]interface{}) {
	pw.AppendNewLineOrDie(`name: netperf-cli`)
	pw.AppendNewLineOrDie(`image: kkourt/netperf`)
	pw.AppendNewLineOrDie(`command: ["netperf"]`)
}

func (cnf *NetperfConf) writeCliBaseArgs(pw *utils.PrefixWriter, params map[string]interface{}) {
	serverIP, ok := params["serverIP"]
	if !ok {
		panic("serverIP undefined")
	}

	pw.AppendNewLineOrDie(fmt.Sprintf(`"-l", "%d", # timeout`, cnf.Timeout))
	pw.AppendNewLineOrDie(`"-j", # enable additional statistics`)
	pw.AppendNewLineOrDie(fmt.Sprintf(`"-H", "%v",`, serverIP))
	pw.AppendNewLineOrDie(fmt.Sprintf(`"-t", "%s", # testname`, cnf.TestName))
}

// WriteCliContainerYaml writes the client yaml
func (cnf *NetperfRRConf) WriteCliContainerYaml(pw *utils.PrefixWriter, params map[string]interface{}) {
	cnf.writeCliStart(pw, params)
	pw.AppendNewLineOrDie(`args : [`)
	pw.PushPrefix("    ")
	cnf.writeCliBaseArgs(pw, params)
	pw.AppendNewLineOrDie(`"--",`)
	pw.AppendNewLineOrDie(fmt.Sprintf(`"-P", "%d", # data connection port`, cnf.DataPort))
	pw.AppendNewLineOrDie(`"-k", "THROUGHPUT,THROUGHPUT_UNITS,P50_LATENCY,P99_LATENCY,REQUEST_SIZE,RESPONSE_SIZE",`)
	pw.PopPrefix()
	pw.AppendNewLineOrDie(`]`)
}

// WriteSrvPortsYaml writes the ports part of yaml (e.g., for services)
func (cnf *NetperfConf) WriteSrvPortsYaml(pw *utils.PrefixWriter, params map[string]interface{}) {
	pw.AppendNewLineOrDie(`- name: netperf-ctl`)
	pw.AppendNewLineOrDie(`  protocol: TCP`)
	pw.AppendNewLineOrDie(`  port: 12865`)
	pw.AppendNewLineOrDie(`  targetPort: 12865`)
	pw.AppendNewLineOrDie(`- name: netperf-data`)
	pw.AppendNewLineOrDie(`  protocol: TCP`)
	pw.AppendNewLineOrDie(fmt.Sprintf(`  port: %d`, cnf.DataPort))
	pw.AppendNewLineOrDie(fmt.Sprintf(`  targetPort: %d`, cnf.DataPort))
}
