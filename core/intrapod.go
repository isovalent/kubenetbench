package core

import (
	"fmt"
	"log"
	"os"
	"text/template"
	"time"

	"../utils"
)

// InterpodSt is the necessary state for executing an intrapod (pod-to-pod) benchmark
type IntrapodSt struct {
	Runctx      *RunCtx
	NetperfConf *NetperfConf // NB: at some point we might want to replace this with an interface
	Policy      string
}

var IntrapodSrvTemplate = template.Must(template.New("srv").Parse(`apiVersion: v1
kind: Pod
metadata:
  name: kubenetbench-{{.runID}}-srv
  labels : {
    runid: {{.runID}},
    role: srv,
  }
spec:
  containers:
  - {{.srvContainer}}
`))

func (s *IntrapodSt) genSrvYaml() (string, error) {
	vals := map[string]interface{}{
		"runID":        s.Runctx.id,
		"srvContainer": "{{template \"netperf\"}}",
	}

	templates := map[string]*template.Template{
		"netperf": netperfSrvYaml(),
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

var IntrapodPortPolicyTemplate = template.Must(template.New("policy").Parse(`apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: kubenetbench-{{.runID}}-policy
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

func (s *IntrapodSt) genPortPolicyYaml() string {
	m := map[string]interface{}{
		"runID":    s.Runctx.id,
		"dataPort": s.NetperfConf.DataPort,
	}

	yaml := fmt.Sprintf("%s/port-policy.yaml", s.Runctx.dir)
	if !s.Runctx.quiet {
		log.Printf("Generating %s", yaml)
	}
	f, err := os.Create(yaml)
	if err != nil {
		log.Fatal(err)
	}
	IntrapodPortPolicyTemplate.Execute(f, m)
	f.Close()
	return yaml
}

var IntrapodCliTemplate = template.Must(template.New("cli").Parse(`apiVersion: v1
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

func (s *IntrapodSt) getCliYaml(serverIP string) (string, error) {
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
		"timeout":      s.NetperfConf.Timeout,
		"serverIP":     serverIP,
		"dataPort":     s.NetperfConf.DataPort,
		"cliContainer": "{{template \"netperf\"}}",
	}

	templates := map[string]*template.Template{
		"netperf": netperfCliYaml(),
	}

	utils.RenderTemplate(IntrapodCliTemplate, vals, templates, f)
	f.Close()
	return yaml, nil
}

func (s IntrapodSt) Execute() error {
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

	// start policy if specified
	if s.Policy == "port" {
		policyYamlFname := s.genPortPolicyYaml()
		err := s.Runctx.KubeApply(policyYamlFname)
		if err != nil {
			return fmt.Errorf("failed to apply policy: %w", err)
		}
	}

	// start netperf client (netperf)
	cliYamlFname, err := s.getCliYaml(srvIP)
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
	time.Sleep(time.Duration(10+s.NetperfConf.Timeout) * time.Second)

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
