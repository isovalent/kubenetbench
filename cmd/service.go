package cmd

import (
	"fmt"
	"log"
	"os"
	"text/template"
	"time"

	"github.com/spf13/cobra"
)

var serviceTypeArg string

type serviceSt struct {
	netperfCnf  netperfConf
	runctx      *runCtx
	serviceType string
}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "service network benchmark",
	Run: func(cmd *cobra.Command, args []string) {

		if serviceTypeArg != "ClusterIP" {
			log.Fatal("invalid policy: ", serviceTypeArg)
		}

		st := serviceSt{
			netperfCnf:  netperfConfDefault(),
			runctx:      newRunCtx(),
			serviceType: serviceTypeArg,
		}
		st.execute()
	},
}

func init() {
	serviceCmd.Flags().StringVar(&serviceTypeArg, "type", "ClusterIP", "service type (ClusterIP)")
}

var serviceYamlTemplate = template.Must(template.New("service").Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: netperf-{{.runID}}-deployment
  labels : {
    runid: {{.runID}},
    role: srv,
  }
spec:
  replicas: 1
  selector:
    matchLabels:
      runid: netperf-{{.runID}}
      role: srv
  template:
    metadata:
      labels : {
        runid: netperf-{{.runID}},
        role: srv,
      }
    spec:
      containers:
      - name: netperf
        image: kkourt/netperf
        command: ["netserver"]
        # -D: dont daemonize
        args: ["-D"]
---
apiVersion: v1
kind: Service
metadata:
  name: netperf-{{.runID}}-service
  labels : {
    runid: {{.runID}},
    role: srv,
  }
spec:
  selector:
    runid: netperf-{{.runID}}
    role: srv
  ports:
    - name: netperf-ctl
      protocol: TCP
      port: 12865
      targetPort: 12865
    - name: netperf-data
      protocol: TCP
      port: {{.dataPort}}
      targetPort: {{.dataPort}}
`))

func (s *serviceSt) genServiceYaml() string {
	m := map[string]interface{}{
		"runID":    s.runctx.id,
		"dataPort": s.netperfCnf.dataPort,
	}

	yaml := fmt.Sprintf("%s/netserv.yaml", s.runctx.dir)
	if !s.runctx.quiet {
		log.Printf("Generating %s", yaml)
	}
	f, err := os.Create(yaml)
	if err != nil {
		log.Fatal(err)
	}
	serviceYamlTemplate.Execute(f, m)
	f.Close()
	return yaml
}

func (s *serviceSt) execute() {
	defer func() {
		if s.runctx.cleanup {
			// FIXME: this does not work because we call functions that
			// call log.Fatal() which calls exit() which does not run the
			// deferred operations
			delPodsCmd := fmt.Sprintf("kubectl delete deployment,service,networkpolicy -l \"runid=%s\"", s.runctx.id)
			s.runctx.ExecCmd(delPodsCmd)
		}
	}()

	// start netperf service (netserver)
	serviceYaml := s.genServiceYaml()
	serviceCmd := fmt.Sprintf("kubectl apply -f %s", serviceYaml)
	s.runctx.ExecCmd(serviceCmd)

	srvSelector := fmt.Sprintf("runid=%s,role=srv", s.runctx.id)
	srvIP, err := s.runctx.KubeGetServiceIP(srvSelector)
	if err != nil {
		log.Fatal(err)
	}

	// start netperf client (netperf)
	cliYaml := s.netperfCnf.genCliYaml(s.runctx, srvIP)
	cliCmd := fmt.Sprintf("kubectl apply -f %s", cliYaml)
	s.runctx.ExecCmd(cliCmd)

	// sleep the duration of the benchmark plus 10s
	time.Sleep(time.Duration(10+s.netperfCnf.timeout) * time.Second)

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
