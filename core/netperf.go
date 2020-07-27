package core

import (
	"text/template"
)

type NetperfConf struct {
	Timeout  int
	DataPort uint16
}

func NetperfConfDefault() *NetperfConf {
	return &NetperfConf{
		Timeout:  10,
		DataPort: 8000,
		// ctl_port:  12865,
	}
}

var netperfSrvYamlTempl_ = template.Must(template.New("netserver").Parse(`name: netperf
image: kkourt/netperf
command: ["netserver"]
# -D: dont daemonize
args: ["-D"]
`))

func netperfSrvYaml() *template.Template {
	return netperfSrvYamlTempl_
}

var netperfCliYamlTempl_ = template.Must(template.New("netperf").Parse(`name: netperf
image: kkourt/netperf
command: ["netperf"]
args: [
    "-l", "{{.timeout}}",     # timeout
    "-j",                     # enable additional statistics
    "-H", "{{.serverIP}}",    # server IP
    "-t", "tcp_rr",           # test name
    "--",                     # test-specific arguments
    "-P", "{{.dataPort}}",    # data connection port
    # additional metrics to record
    "-k", "THROUGHPUT,THROUGHPUT_UNITS,P50_LATENCY,P99_LATENCY,REQUEST_SIZE,RESPONSE_SIZE"
]
`))

func netperfCliYaml() *template.Template {
	return netperfCliYamlTempl_
}
