package core

import (
	"fmt"
	"log"
	"os"
	"text/template"
	"time"

	"../utils"
)

type ServiceSt struct {
	Runctx      *RunCtx
	ServiceType string
}

var ServiceYamlTemplate = template.Must(template.New("service").Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubenetbench-{{.runID}}-deployment
  labels : {
    runid: {{.runID}},
    role: srv,
  }
spec:
  replicas: 1
  selector:
    matchLabels:
      runid: kubenetbench-{{.runID}}
      role: srv
  template:
    metadata:
      labels : {
        runid: kubenetbench-{{.runID}},
        role: srv,
      }
    spec:
      containers:
      - {{.srvContainer}}
---
apiVersion: v1
kind: Service
metadata:
  name: kubenetbench-{{.runID}}-service
  labels : {
    runid: {{.runID}},
    role: srv,
  }
spec:
  selector:
    runid: kubenetbench-{{.runID}}
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

func (s *ServiceSt) genSrvYaml() (string, error) {
	vals := map[string]interface{}{
		"runID":        s.Runctx.id,
		"srvContainer": "{{template \"netperf\"}}",
	}

	templates := map[string]utils.PrefixRenderer{
		"netperf": s.Runctx.benchmark.WriteSrvYaml,
	}

	yaml := fmt.Sprintf("%s/netserv.yaml", s.Runctx.dir)
	if !s.Runctx.quiet {
		log.Printf("Generating %s", yaml)
	}
	f, err := os.Create(yaml)
	if err != nil {
		return "", err
	}
	utils.RenderTemplate(IntrapodSrvTemplate, vals, templates, f)
	f.Close()
	return yaml, nil
}

var ServiceCliTemplate = template.Must(template.New("cli").Parse(`apiVersion: v1
kind: Pod
metadata:
  name: kubenetbench-{{.runID}}-cli
  labels : {
     runid: {{.runID}},
     role: cli,
  }
spec:
  restartPolicy: Never
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
		"cliContainer": "{{template \"netperf\"}}",
	}

	templates := map[string]utils.PrefixRenderer{
		"netperf": s.Runctx.benchmark.WriteCliYaml,
	}

	utils.RenderTemplate(IntrapodCliTemplate, vals, templates, f)
	f.Close()
	return yaml, nil
}

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

	srvSelector := fmt.Sprintf("runid=%s,role=srv", s.Runctx.id)

	defer func() {
		// attempt to save server logs
		s.Runctx.KubeSaveLogs(srvSelector, fmt.Sprintf("%s/srv.log", s.Runctx.dir))

		// FIXME: this does not work because we call functions that
		// call log.Fatal() which calls exit() which does not run the
		// deferred operations
		s.Runctx.KubeCleanup()
	}()

	// get server pod IP
	time.Sleep(2 * time.Second)
	srvIP, err := s.Runctx.KubeGetPodIP(srvSelector, 10, 2*time.Second)
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

	cliSelector := fmt.Sprintf("runid=%s,role=cli", s.Runctx.id)
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
			log.Printf("Client phase: %s", cliPhase)
		}

		if cliPhase == "Succeeded" || cliPhase == "Failed" {
			break
		}
		time.Sleep(2 * time.Second)
	}

	return nil
}
