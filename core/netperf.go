package core

import (
	"fmt"
	"strings"

	"../utils"
)

// NetperfConf base netperf configuration
type NetperfConf struct {
	Timeout    int
	DataPort   uint16
	TestName   string
	CliCommand string
}

// NetperfConfDefault returns a NetperfConf with the default values
func NetperfConfDefault(testname string) NetperfConf {
	return NetperfConf{
		Timeout:    60,
		DataPort:   8000,
		CliCommand: "netperf",
		TestName:   testname,
	}
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
	pw.AppendNewLineOrDie(`image: kkourt/kubenetbench`)
	pw.AppendNewLineOrDie(`command: ["netserver"]`)
	pw.AppendNewLineOrDie(`args : [`)
	pw.PushPrefix("    ")
	pw.AppendNewLineOrDie(`"-D", # dont daemonize`)
	pw.PopPrefix()
	pw.AppendNewLineOrDie(`]`)
}

// Additional options, we might want to add
// -I <optionspec>
//     This option enables the calculation of confidence intervals and sets the
//     confidence and width parameters with the first half of the optionspec being
//     either 99 or 95 for 99% or 95% confidence respectively. The second value of
//     the optionspec specifies the width of the desired confidence interval.
//
// -T <optionspec>
//       This option controls the CPU, and probably by extension memory, affinity of netperf and/or netserver.

// WriteCliContainerYaml writes the client yaml
func (cnf *NetperfRRConf) WriteCliContainerYaml(pw *utils.PrefixWriter, params map[string]interface{}) {
	serverIP, ok := params["serverIP"]
	if !ok {
		panic("serverIP undefined")
	}
	outputFields := []string{
		"THROUGHPUT",
		"THROUGHPUT_UNITS",
		"TRANSACTION_RATE",
		"P50_LATENCY",
		"P90_LATENCY",
		"RT_LATENCY",
		"MEAN_LATENCY",
		"STDEV_LATENCY",
		"REQUEST_SIZE",
		"RESPONSE_SIZE",
		// "DIRECTION",
		// "LOCAL_CPU_BIND",
		// "REMOTE_CPU_BIND",
		"LOCAL_TRANSPORT_RETRANS",
		"REMOTE_TRANSPORT_RETRANS",
	}
	pw.AppendNewLineOrDie(`name: netperf-cli`)
	pw.AppendNewLineOrDie(`image: kkourt/kubenetbench`)
	pw.AppendNewLineOrDie(fmt.Sprintf(`command: ["%s"]`, cnf.CliCommand))
	pw.AppendNewLineOrDie(`args : [`)
	pw.PushPrefix("    ")
	pw.AppendNewLineOrDie(fmt.Sprintf(`"-l", "%d", # timeout`, cnf.Timeout))
	pw.AppendNewLineOrDie(`"-j", # enable additional statistics`)
	pw.AppendNewLineOrDie(fmt.Sprintf(`"-H", "%v",`, serverIP))
	pw.AppendNewLineOrDie(fmt.Sprintf(`"-t", "%s", # testname`, cnf.TestName))
	pw.AppendNewLineOrDie(`"--",`)
	pw.AppendNewLineOrDie(fmt.Sprintf(`"-P", ",%d", # data connection port`, cnf.DataPort))
	// -D seems to kill the performance for high queue depths, so don't use it
	// pw.AppendNewLineOrDie(`"-D",# no delay`)
	pw.AppendNewLineOrDie(fmt.Sprintf(`"-k", "%s",`, strings.Join(outputFields, ",")))
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
