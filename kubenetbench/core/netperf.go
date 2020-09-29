package core

import (
	"fmt"
	"strings"

	"github.com/cilium/kubenetbench/utils"
)

// NetperfConf base netperf configuration
type NetperfConf struct {
	Timeout       int
	DataPort      uint16
	TestName      string
	CliCommand    string
	PreArgs       []string
	MoreArgs      []string
	MoreBenchArgs []string
}

// NetperfConfDefault returns a NetperfConf with the default values
func NetperfConfDefault(testname string, args []string, benchArgs []string) NetperfConf {
	return NetperfConf{
		Timeout:       60,
		DataPort:      8000,
		CliCommand:    "netperf",
		TestName:      testname,
		MoreArgs:      args,
		MoreBenchArgs: benchArgs,
	}
}

// GetTimeout returns the benchmark timeout
func (cnf *NetperfConf) GetTimeout() int {
	return cnf.Timeout
}

// WriteSrvContainerYaml writes the server yaml
func (cnf *NetperfConf) WriteSrvContainerYaml(pw *utils.PrefixWriter, params map[string]interface{}) {
	pw.AppendNewLineOrDie(`name: netperf-srv`)
	pw.AppendNewLineOrDie(`image: cilium/kubenetbench`)
	pw.AppendNewLineOrDie(`command: ["netserver"]`)
	pw.AppendNewLineOrDie(`args : [`)
	pw.PushPrefix("    ")
	pw.AppendNewLineOrDie(`"-D", # dont daemonize`)
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

/**
 * RR
 */

// NetperfRRConf RR netperf configuration
type NetperfRRConf struct {
	NetperfConf
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

func netperfOutFieldsCommon() []string {
	return []string{
		"THROUGHPUT",
		"THROUGHPUT_UNITS",
		"THROUGHPUT_CONFID",
		"PROTOCOL",
		"ELAPSED_TIME",
		//
		"LOCAL_SEND_CALLS",
		"LOCAL_BYTES_PER_SEND",
		"LOCAL_RECV_CALLS",
		"LOCAL_BYTES_PER_RECV",
		"REMOTE_SEND_CALLS",
		"REMOTE_BYTES_PER_SEND",
		"REMOTE_RECV_CALLS",
		"REMOTE_BYTES_PER_RECV",
		//
		"LOCAL_SYSNAME",
		"LOCAL_RELEASE",
		"LOCAL_VERSION",
		"LOCAL_MACHINE",
		"REMOTEL_SYSNAME",
		"REMOTEL_RELEASE",
		"REMOTEL_VERSION",
		"REMOTEL_MACHINE",
		//
		"COMMAND_LINE",
		// "DIRECTION",
		// "LOCAL_CPU_BIND",
		// "REMOTE_CPU_BIND",
		"LOCAL_TRANSPORT_RETRANS",
		"REMOTE_TRANSPORT_RETRANS",
	}
}

// WriteCliContainerYaml writes the client yaml
func (cnf *NetperfRRConf) WriteCliContainerYaml(pw *utils.PrefixWriter, params map[string]interface{}) {
	serverIP, ok := params["serverIP"]
	if !ok {
		panic("serverIP undefined")
	}

	outputFields := append(
		netperfOutFieldsCommon(),
		"TRANSACTION_RATE",
		"P50_LATENCY",
		"P90_LATENCY",
		"RT_LATENCY",
		"MEAN_LATENCY",
		"STDEV_LATENCY",
		"REQUEST_SIZE",
		"RESPONSE_SIZE",
		"BURST_SIZE",
	)

	pw.AppendNewLineOrDie(`name: netperf-cli`)
	pw.AppendNewLineOrDie(`image: cilium/kubenetbench`)
	pw.AppendNewLineOrDie(fmt.Sprintf(`command: ["%s"]`, cnf.CliCommand))
	pw.AppendNewLineOrDie(`args : [`)
	pw.PushPrefix("    ")
	if len(cnf.PreArgs) > 0 {
		pw.AppendNewLineOrDie("# initial args")
		for _, arg := range cnf.PreArgs {
			pw.AppendNewLineOrDie(fmt.Sprintf(`"%s",`, arg))
		}
	}
	pw.AppendNewLineOrDie(fmt.Sprintf(`"-l", "%d", # timeout`, cnf.Timeout))
	pw.AppendNewLineOrDie(`"-j", # enable additional statistics`)
	pw.AppendNewLineOrDie(fmt.Sprintf(`"-H", "%v",`, serverIP))
	pw.AppendNewLineOrDie(fmt.Sprintf(`"-t", "%s", # testname`, cnf.TestName))
	if len(cnf.MoreArgs) > 0 {
		pw.AppendNewLineOrDie("# Additional args")
		for _, arg := range cnf.MoreArgs {
			pw.AppendNewLineOrDie(fmt.Sprintf(`"%s",`, arg))
		}
	}

	pw.AppendNewLineOrDie(`"--",`)
	pw.AppendNewLineOrDie("# Benchmark args")
	pw.AppendNewLineOrDie(fmt.Sprintf(`"-P", ",%d", # data connection port`, cnf.DataPort))
	// -D seems to kill the performance for high queue depths, so don't use it
	// pw.AppendNewLineOrDie(`"-D",# no delay`)
	pw.AppendNewLineOrDie(fmt.Sprintf(`"-k", "%s",`, strings.Join(outputFields, ",")))
	if len(cnf.MoreBenchArgs) > 0 {
		pw.AppendNewLineOrDie("# Additional test-specific args")
		for _, arg := range cnf.MoreBenchArgs {
			pw.AppendNewLineOrDie(fmt.Sprintf(`"%s",`, arg))
		}
	}
	pw.PopPrefix()
	pw.AppendNewLineOrDie(`]`)
}

/**
 * STREAM
 */

// NetperfStreamConf STREAM netperf configuration
type NetperfStreamConf struct {
	NetperfConf
}

// WriteCliContainerYaml writes the client yaml
func (cnf *NetperfStreamConf) WriteCliContainerYaml(pw *utils.PrefixWriter, params map[string]interface{}) {
	serverIP, ok := params["serverIP"]
	if !ok {
		panic("serverIP undefined")
	}
	outputFields := []string{
		"THROUGHPUT",
		"THROUGHPUT_UNITS",
		"THROUGHPUT_CONFID",
		"LOCAL_SEND_SIZE",
		"LOCAL_RECV_SIZE",
		"REMOTE_SEND_SIZE",
		"REMOTE_RECV_SIZE",
		"PROTOCOL",
		"LOCAL_SEND_CALLS",
		"LOCAL_BYTES_PER_SEND",
		"LOCAL_RECV_CALLS",
		"LOCAL_BYTES_PER_RECV",
		"REMOTE_SEND_CALLS",
		"REMOTE_BYTES_PER_SEND",
		"REMOTE_RECV_CALLS",
		"REMOTE_BYTES_PER_RECV",
		//
		"LOCAL_SEND_THROUGHPUT",
		"LOCAL_RECV_THROUGHPUT",
		"REMOTE_SEND_THROUGHPUT",
		"REMOTE_RECV_THROUGHPUT",
		//
		"LOCAL_SYSNAME",
		"LOCAL_RELEASE",
		"LOCAL_VERSION",
		"LOCAL_MACHINE",
		"REMOTEL_SYSNAME",
		"REMOTEL_RELEASE",
		"REMOTEL_VERSION",
		"REMOTEL_MACHINE",
		//
		"COMMAND_LINE",
		// "DIRECTION",
		// "LOCAL_CPU_BIND",
		// "REMOTE_CPU_BIND",
		"LOCAL_TRANSPORT_RETRANS",
		"REMOTE_TRANSPORT_RETRANS",
	}
	pw.AppendNewLineOrDie(`name: netperf-cli`)
	pw.AppendNewLineOrDie(`image: cilium/kubenetbench`)
	pw.AppendNewLineOrDie(fmt.Sprintf(`command: ["%s"]`, cnf.CliCommand))
	pw.AppendNewLineOrDie(`args : [`)
	pw.PushPrefix("    ")
	if len(cnf.PreArgs) > 0 {
		pw.AppendNewLineOrDie("# initial args")
		for _, arg := range cnf.PreArgs {
			pw.AppendNewLineOrDie(fmt.Sprintf(`"%s",`, arg))
		}
	}
	pw.AppendNewLineOrDie(fmt.Sprintf(`"-l", "%d", # timeout`, cnf.Timeout))
	pw.AppendNewLineOrDie(`"-j", # enable additional statistics`)
	pw.AppendNewLineOrDie(fmt.Sprintf(`"-H", "%v",`, serverIP))
	pw.AppendNewLineOrDie(fmt.Sprintf(`"-t", "%s", # testname`, cnf.TestName))
	if len(cnf.MoreArgs) > 0 {
		pw.AppendNewLineOrDie("# Additional args")
		for _, arg := range cnf.MoreArgs {
			pw.AppendNewLineOrDie(fmt.Sprintf(`"%s",`, arg))
		}
	}

	pw.AppendNewLineOrDie(`"--",`)
	pw.AppendNewLineOrDie("# Benchmark args")
	pw.AppendNewLineOrDie(fmt.Sprintf(`"-P", ",%d", # data connection port`, cnf.DataPort))
	// -D seems to kill the performance for high queue depths, so don't use it
	// pw.AppendNewLineOrDie(`"-D",# no delay`)
	pw.AppendNewLineOrDie(fmt.Sprintf(`"-k", "%s",`, strings.Join(outputFields, ",")))

	// netperf seems to be setting SO_DONTROUTE for udp_stream, which might
	// not work in many setups. -R 1 disables this.
	if cnf.TestName == "udp_stream" {
		pw.AppendNewLineOrDie(`"-R", "1"`)
	}

	if len(cnf.MoreBenchArgs) > 0 {
		pw.AppendNewLineOrDie("# Additional test-specific args")
		for _, arg := range cnf.MoreBenchArgs {
			pw.AppendNewLineOrDie(fmt.Sprintf(`"%s",`, arg))
		}
	}
	pw.PopPrefix()
	pw.AppendNewLineOrDie(`]`)
}
