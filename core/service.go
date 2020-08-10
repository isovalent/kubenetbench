package core

import (
	"fmt"
	"log"
	"os"
	"text/template"
	"time"

	"../utils"
)

// ServiceSt is the state for the service run
type ServiceSt struct {
	Runctx      *RunCtx
	ServiceType string
}

var serviceYamlTemplate = template.Must(template.New("service").Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubenetbench-{{.runID}}-deployment
  labels : {
    kubenetbench-runid: {{.runID}},
    role: srv,
  }
spec:
  replicas: 1
  selector:
    matchLabels:
      kubenetbench-runid: {{.runID}}
      role: srv
  template:
    metadata:
      labels : {
        kubenetbench-runid: {{.runID}},
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
		"runID":        s.Runctx.id,
		"srvContainer": "{{template \"netperfContainer\"}}",
		"srvPorts":     "{{template \"netperfPorts\"}}",
		"srvAffinity":  "{{template \"srvAffinity\"}}",
	}

	templates := map[string]utils.PrefixRenderer{
		"netperfContainer": s.Runctx.benchmark.WriteSrvContainerYaml,
		"netperfPorts":     s.Runctx.benchmark.WriteSrvPortsYaml,
		"srvAffinity":      s.Runctx.srvAffinityWrite,
	}

	yaml := fmt.Sprintf("%s/netserv.yaml", s.Runctx.dir)
	if !s.Runctx.quiet {
		log.Printf("Generating %s", yaml)
	}
	f, err := os.Create(yaml)
	if err != nil {
		return "", err
	}
	utils.RenderTemplate(serviceYamlTemplate, vals, templates, f)
	f.Close()
	return yaml, nil
}

var serviceCliTemplate = template.Must(template.New("cli").Parse(`apiVersion: v1
kind: Pod
metadata:
  name: kubenetbench-{{.runID}}-cli
  labels : {
     kubenetbench-runid: {{.runID}},
     role: cli,
  }
spec:
  restartPolicy: Never
  {{.cliAffinity}}
  containers:
  - {{.cliContainer}}
`))

func (s *ServiceSt) genCliYaml(serverIP string) (string, error) {
	yaml := fmt.Sprintf("%s/client.yaml", s.Runctx.dir)
	if !s.Runctx.quiet {
		log.Printf("Generating %s", yaml)
	}
	f, err := os.Create(yaml)
	if err != nil {
		return "", err
	}

	vals := map[string]interface{}{
		"runID":        s.Runctx.id,
		"serverIP":     serverIP,
		"cliContainer": "{{template \"netperfContainer\"}}",
		"cliAffinity":  "{{template \"cliAffinity\"}}",
	}

	templates := map[string]utils.PrefixRenderer{
		"netperfContainer": s.Runctx.benchmark.WriteCliContainerYaml,
		"cliAffinity":      s.Runctx.cliAffinityWrite,
	}

	utils.RenderTemplate(serviceCliTemplate, vals, templates, f)
	f.Close()
	return yaml, nil
}

// Execute service run
func (s ServiceSt) Execute() error {
	// start server pod (netserver)
	srvYamlFname, err := s.genSrvYaml()
	if err != nil {
		return err
	}
	err = s.Runctx.KubeApply(srvYamlFname)
	if err != nil {
		return err
	}

	srvSelector := fmt.Sprintf("kubenetbench-runid=%s,role=srv", s.Runctx.id)

	defer func() {
		// attempt to save server logs
		s.Runctx.KubeSaveLogs(srvSelector, fmt.Sprintf("%s/srv.log", s.Runctx.dir))

		// FIXME: this does not work because we call functions that
		// call log.Fatal() which calls exit() which does not run the
		// deferred operations
		s.Runctx.KubeCleanup()
	}()

	// get service IP
	time.Sleep(2 * time.Second)
	srvIP, err := s.Runctx.KubeGetServiceIP(srvSelector, 10, 2*time.Second)
	if err != nil {
		return err
	}
	if !s.Runctx.quiet {
		log.Printf("server_ip=%s", srvIP)
	}

	// start netperf client (netperf)
	cliYamlFname, err := s.genCliYaml(srvIP)
	if err != nil {
		return err
	}

	err = s.Runctx.KubeApply(cliYamlFname)
	if err != nil {
		return fmt.Errorf("failed to initiate client: %w", err)
	}

	cliSelector := fmt.Sprintf("kubenetbench-runid=%s,role=cli", s.Runctx.id)
	// attempt to save client logs
	defer s.Runctx.KubeSaveLogs(cliSelector, fmt.Sprintf("%s/cli.log", s.Runctx.dir))

	// sleep the duration of the benchmark plus 10s
	time.Sleep(time.Duration(10+s.Runctx.benchmark.GetTimeout()) * time.Second)

	var cliPhase string
	for {
		cliPhase, err = s.Runctx.KubeGetPodPhase(cliSelector)
		if err != nil {
			return err
		}
		if !s.Runctx.quiet {
			log.Printf("client phase: %s", cliPhase)
		}

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
