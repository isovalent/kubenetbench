package core

import (
	"fmt"
	"log"
	"os"
	"text/template"
	"time"

	"github.com/kkourt/kubenetbench/utils"
)

// ServiceSt is the state for the service run
type ServiceSt struct {
	RunBenchCtx *RunBenchCtx
	ServiceType string
}

var serviceYamlTemplate = template.Must(template.New("service").Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: knb-deployment
  labels : {
    {{.runLabel}},
    role: srv,
  }
spec:
  replicas: 1
  selector:
    matchLabels:
      {{.runLabel}},
      role: srv
  template:
    metadata:
      labels : {
        {{.runLabel}},
        role: srv,
      }
    spec:
      {{.srvAffinity}}
      containers:
      - {{.srvContainer}}
---
apiVersion: v1
kind: Service
metadata:
  name: kubenetbench-{{.runID}}-service
  labels : {
    kubenetbench-runid: {{.runID}},
    role: srv,
  }
spec:
  selector:
    kubenetbench-runid: {{.runID}}
    role: srv
  ports:
    {{.srvPorts}}
`))

func (s *ServiceSt) genSrvYaml() (string, error) {
	vals := map[string]interface{}{
		"runLabel":     s.RunBenchCtx.getRunLabel(": "),
		"srvContainer": "{{template \"netperfContainer\"}}",
		"srvPorts":     "{{template \"netperfPorts\"}}",
		"srvAffinity":  "{{template \"srvAffinity\"}}",
	}

	templates := map[string]utils.PrefixRenderer{
		"netperfContainer": s.RunBenchCtx.benchmark.WriteSrvContainerYaml,
		"netperfPorts":     s.RunBenchCtx.benchmark.WriteSrvPortsYaml,
		"srvAffinity":      s.RunBenchCtx.srvAffinityWrite,
	}

	yaml := fmt.Sprintf("%s/netserv.yaml", s.RunBenchCtx.getDir())
	log.Printf("Generating %s", yaml)
	f, err := os.Create(yaml)
	if err != nil {
		return "", err
	}
	utils.RenderTemplate(serviceYamlTemplate, vals, templates, f)
	f.Close()
	return yaml, nil
}

func (s *ServiceSt) genCliYaml(serverIP string) (string, error) {
	return s.RunBenchCtx.genCliYaml(serverIP)
}

// Execute service run
func (s ServiceSt) Execute() error {
	// start server pod (netserver)
	srvYamlFname, err := s.genSrvYaml()
	if err != nil {
		return err
	}
	err = s.RunBenchCtx.KubeApply(srvYamlFname)
	if err != nil {
		return err
	}

	srvSelector := fmt.Sprintf("%s,role=srv", s.RunBenchCtx.getRunLabel("="))

	defer func() {
		// attempt to save server logs
		s.RunBenchCtx.KubeSaveLogs(srvSelector, fmt.Sprintf("%s/srv.log", s.RunBenchCtx.getDir()))

		// FIXME: this does not work because we call functions that
		// call log.Fatal() which calls exit() which does not run the
		// deferred operations
		s.RunBenchCtx.KubeCleanup()
	}()

	// get service IP
	time.Sleep(2 * time.Second)
	srvIP, err := s.RunBenchCtx.KubeGetServiceIP(srvSelector, 10, 2*time.Second)
	if err != nil {
		return err
	}
	log.Printf("server_ip=%s", srvIP)

	// start netperf client (netperf)
	cliYamlFname, err := s.genCliYaml(srvIP)
	if err != nil {
		return err
	}

	err = s.RunBenchCtx.KubeApply(cliYamlFname)
	if err != nil {
		return fmt.Errorf("failed to initiate client: %w", err)
	}

	cliSelector := fmt.Sprintf("%s,role=cli", s.RunBenchCtx.getRunLabel("="))
	// attempt to save client logs
	defer s.RunBenchCtx.KubeSaveLogs(cliSelector, fmt.Sprintf("%s/cli.log", s.RunBenchCtx.getDir()))

	// sleep the duration of the benchmark plus 10s
	time.Sleep(time.Duration(10+s.RunBenchCtx.benchmark.GetTimeout()) * time.Second)

	var cliPhase string
	for {
		cliPhase, err = s.RunBenchCtx.KubeGetPodPhase(cliSelector)
		if err != nil {
			return err
		}
		log.Printf("client phase: %s", cliPhase)

		if cliPhase == "Succeeded" {
			return nil
		}
		if cliPhase == "Failed" {
			return fmt.Errorf("client execution failed")
		}
		time.Sleep(10 * time.Second)
	}

	return nil
}
