package cmd

import (
	"fmt"
	"log"
	"os"
	"text/template"
)

type netperfConf struct {
	timeout  int
	dataPort uint16
}

// NB: eventually we might want to put this in its own file, with and declare
// netperf specific flags and also add other "backends" like ipperf.  It would
// be good if we had a way to define the netperf configuration independelty and
// embed it into the YAML files we generate. https://cuelang.org/ might be a
// good way to achieve this.
func netperfConfDefault() netperfConf {
	return netperfConf{
		timeout:  10,
		dataPort: 8000,
		// ctl_port:  12865,
	}
}

var cliYamlTemplate = template.Must(template.New("cli").Parse(`apiVersion: v1
kind: Pod
metadata:
  name: netperf-{{.runID}}-cli
  labels : {
     runid: {{.runID}},
     role: cli,
  }
spec:
  restartPolicy: Never
  containers:
  - name: netperf
    image: kkourt/netperf
    command: ["netperf"]
    args: [
        "-l", "{{.timeout}}",     # timeout
        "-j",                     # enable additional statistics
        "-H", "{{.serverIP}}",    # server IP
        "-t", "tcp_rr",           # test name
        "--",                     # test specific arguemnts
        "-P", "{{.dataPort}}",    # data connection port
        # additional metrics to record
        "-k", "THROUGHPUT,THROUGHPUT_UNITS,P50_LATENCY,P99_LATENCY,REQUEST_SIZE,RESPONSE_SIZE"
    ]
`))

func (np *netperfConf) genCliYaml(run *runCtx, serverIP string) string {
	m := map[string]interface{}{
		"runID":    run.id,
		"timeout":  np.timeout,
		"serverIP": serverIP,
		"dataPort": np.dataPort,
	}

	yaml := fmt.Sprintf("%s/client.yaml", run.dir)
	if !run.quiet {
		log.Printf("Generating %s", yaml)
	}
	f, err := os.Create(yaml)
	if err != nil {
		log.Fatal(err)
	}
	cliYamlTemplate.Execute(f, m)
	f.Close()
	return yaml
}
