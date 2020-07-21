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

type state struct {
	NetperfCnf netperfConf
	runctx     *runCtx
	policy     string
}

var policyArg string

var intrapodCmd = &cobra.Command{
	Use:   "intrapod",
	Short: "pod-to-pod network benchmark",
	Run: func(cmd *cobra.Command, args []string) {
		if policyArg != "" && policyArg != "port" {
			log.Fatal("invalid policy: ", policyArg)
		}

		ctx := state{
			NetperfCnf: netperfConf{
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
  name: netperf-{{.runID}}-srv
  labels : {
    runid: {{.runID}},
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

func (s *state) genSrvYaml() string {
	m := map[string]interface{}{
		"runID": s.runctx.id,
	}

	yaml := fmt.Sprintf("%s/netserv.yaml", s.runctx.dir)
	if !s.runctx.quiet {
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

func (s *state) genCliYaml(serverIP string) string {
	m := map[string]interface{}{
		"runID":    s.runctx.id,
		"timeout":  s.NetperfCnf.timeout,
		"serverIP": serverIP,
		"dataPort": s.NetperfCnf.dataPort,
	}

	yaml := fmt.Sprintf("%s/client.yaml", s.runctx.dir)
	if !s.runctx.quiet {
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
  name: netperf-{{.runID}}-policy
  labels : {
     "runid": {{.runID}},
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

func (s *state) genPortPolicyYaml() string {
	m := map[string]interface{}{
		"runID":    s.runctx.id,
		"dataPort": s.NetperfCnf.dataPort,
	}

	yaml := fmt.Sprintf("%s/port-policy.yaml", s.runctx.dir)
	if !s.runctx.quiet {
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

func (s state) execute() {
	// start netperf server (netserver)
	srvYaml := s.genSrvYaml()
	srvCmd := fmt.Sprintf("kubectl apply -f %s", srvYaml)
	s.runctx.ExecCmd(srvCmd)

	defer func() {
		// FIXME: this does not work if there is an error and we exit()
		delPodsCmd := fmt.Sprintf("kubectl delete pod,networkpolicy -l \"runid=%s\"", s.runctx.id)
		s.runctx.ExecCmd(delPodsCmd)
	}()

	// get its IP
	srvSelector := fmt.Sprintf("runid=%s,role=srv", s.runctx.id)
	time.Sleep(2 * time.Second)
	srvIP := s.runctx.KubeGetIP(srvSelector, 10, 2*time.Second)
	if !s.runctx.quiet {
		log.Printf("server_ip=%s", srvIP)
	}

	if s.policy == "port" {
		policyYaml := s.genPortPolicyYaml()
		policyCmd := fmt.Sprintf("kubectl apply -f %s", policyYaml)
		s.runctx.ExecCmd(policyCmd)
	}

	// start netperf client (netperf)
	cliYaml := s.genCliYaml(srvIP)
	cliCmd := fmt.Sprintf("kubectl apply -f %s", cliYaml)
	s.runctx.ExecCmd(cliCmd)

	// sleep the duration of the benchmark plus 10s
	time.Sleep(time.Duration(10+s.NetperfCnf.timeout) * time.Second)

	cliSelector := fmt.Sprintf("runid=%s,role=cli", s.runctx.id)
	var cliPhase string
	for {
		cliPhase = s.runctx.KubeGetPhase(cliSelector)
		if !s.runctx.quiet {
			log.Printf("Client phase: %s", cliPhase)
		}

		if cliPhase == "Succeeded" || cliPhase == "Failed" {
			break
		}
		time.Sleep(2 * time.Second)
	}

	s.runctx.KubeSaveLogs(cliSelector, fmt.Sprintf("%s/cli.log", s.runctx.dir))
	s.runctx.KubeSaveLogs(srvSelector, fmt.Sprintf("%s/srv.log", s.runctx.dir))
}
