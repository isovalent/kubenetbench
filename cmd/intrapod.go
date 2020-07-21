package cmd

import (
	"fmt"
	"log"
	"os"
	"text/template"
	"time"

	"github.com/spf13/cobra"
)

type netperfConf struct {
	timeout  int
	dataPort uint16
}

type Ctx struct {
	netperf_cnf netperfConf
	runctx      *RunCtx
	policy      string
}

var policyArg string

var intrapodCmd = &cobra.Command{
	Use:   "intrapod",
	Short: "pod-to-pod network benchmark",
	Run: func(cmd *cobra.Command, args []string) {
		if policyArg != "" && policyArg != "port" {
			log.Fatal("invalid policy: ", policyArg)
		}

		ctx := Ctx{
			netperf_cnf: netperfConf{
				timeout:  10,
				dataPort: 8000,
				// ctl_port:  12865,
			},
			runctx: newRunCtx(),
			policy: policyArg,
		}
		ctx.execute()
	},
}

func init() {
	intrapodCmd.Flags().StringVar(&policyArg, "policy", "", "isolation policy (empty or \"port\")")
}

var srvYamlTemplate = template.Must(template.New("srv").Parse(`apiVersion: v1
kind: Pod
metadata:
  name: netperf-{{.runId}}-srv
  labels : {
    runid: {{.runId}},
    role: srv,
  }
spec:
  containers:
  - name: netperf
    image: kkourt/netperf
    command: ["netserver"]
    # -D: dont daemonize
    args: ["-D"]
`))

func (c *Ctx) genSrvYaml() string {
	m := map[string]interface{}{
		"runId": c.runctx.id,
	}

	yaml := fmt.Sprintf("%s/netserv.yaml", c.runctx.dir)
	if !c.runctx.quiet {
		log.Printf("Generating %s", yaml)
	}
	f, err := os.Create(yaml)
	if err != nil {
		log.Fatal(err)
	}
	srvYamlTemplate.Execute(f, m)
	f.Close()
	return yaml
}

var cliYamlTemplate = template.Must(template.New("cli").Parse(`apiVersion: v1
kind: Pod
metadata:
  name: netperf-{{.runId}}-cli
  labels : {
     runid: {{.runId}},
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
        "-H", "{{.serverIp}}",    # server IP
        "-t", "tcp_rr",           # test name
        "--",                     # test specific arguemnts
        "-P", "{{.dataPort}}",    # data connection port
        # additional metrics to record
        "-k", "THROUGHPUT,THROUGHPUT_UNITS,P50_LATENCY,P99_LATENCY,REQUEST_SIZE,RESPONSE_SIZE"
    ]
`))

func (c *Ctx) genCliYaml(serverIp string) string {
	m := map[string]interface{}{
		"runId":    c.runctx.id,
		"timeout":  c.netperf_cnf.timeout,
		"serverIp": serverIp,
		"dataPort": c.netperf_cnf.dataPort,
	}

	yaml := fmt.Sprintf("%s/client.yaml", c.runctx.dir)
	if !c.runctx.quiet {
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

var portPolicyYamlTemplate = template.Must(template.New("policy").Parse(`apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: netperf-{{.runId}}-policy
  labels : {
     "runid": {{.runId}},
  }
spec:
  podSelector:
    matchLabels:
      role: srv
  policyTypes:
  - Ingress
  ingress:
  - from:
    ports:
    # control port
    - protocol: TCP
      port: 12865
    # data port
    - protocol: TCP
      port: {{.dataPort}}
`))

func (c *Ctx) genPortPolicyYaml() string {
	m := map[string]interface{}{
		"runId":    c.runctx.id,
		"dataPort": c.netperf_cnf.dataPort,
	}

	yaml := fmt.Sprintf("%s/port-policy.yaml", c.runctx.dir)
	if !c.runctx.quiet {
		log.Printf("Generating %s", yaml)
	}
	f, err := os.Create(yaml)
	if err != nil {
		log.Fatal(err)
	}
	portPolicyYamlTemplate.Execute(f, m)
	f.Close()
	return yaml
}

func (c Ctx) execute() {
	// start netperf server (netserver)
	srvYaml := c.genSrvYaml()
	srvCmd := fmt.Sprintf("kubectl apply -f %s", srvYaml)
	c.runctx.ExecCmd(srvCmd)

	defer func() {
		// FIXME: this does not work if there is an error and we exit()
		delPodsCmd := fmt.Sprintf("kubectl delete pod,networkpolicy -l \"runid=%s\"", c.runctx.id)
		c.runctx.ExecCmd(delPodsCmd)
	}()

	// get its IP
	srvSelector := fmt.Sprintf("runid=%s,role=srv", c.runctx.id)
	time.Sleep(2 * time.Second)
	srvIp := c.runctx.KubeGetIP(srvSelector, 10, 2*time.Second)
	if !c.runctx.quiet {
		log.Printf("server_ip=%s", srvIp)
	}

	if c.policy == "port" {
		policyYaml := c.genPortPolicyYaml()
		policyCmd := fmt.Sprintf("kubectl apply -f %s", policyYaml)
		c.runctx.ExecCmd(policyCmd)
	}

	// start netperf client (netperf)
	cliYaml := c.genCliYaml(srvIp)
	cliCmd := fmt.Sprintf("kubectl apply -f %s", cliYaml)
	c.runctx.ExecCmd(cliCmd)

	// sleep the duration of the benchmark plus 10s
	time.Sleep(time.Duration(10+c.netperf_cnf.timeout) * time.Second)

	cliSelector := fmt.Sprintf("runid=%s,role=cli", c.runctx.id)
	var cliPhase string
	for {
		cliPhase = c.runctx.KubeGetPhase(cliSelector)
		if !c.runctx.quiet {
			log.Printf("Client phase: %s", cliPhase)
		}

		if cliPhase == "Succeeded" || cliPhase == "Failed" {
			break
		}
		time.Sleep(2 * time.Second)
	}

	c.runctx.KubeSaveLogs(cliSelector, fmt.Sprintf("%s/cli.log", c.runctx.dir))
	c.runctx.KubeSaveLogs(srvSelector, fmt.Sprintf("%s/srv.log", c.runctx.dir))
}
