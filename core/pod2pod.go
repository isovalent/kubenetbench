package core

import (
	"fmt"
	"log"
	"os"
	"text/template"
	"time"

	"../utils"
)

// Pod2PodSt is the necessary state for executing an pod-to-pod benchmark
type Pod2PodSt struct {
	Runctx *RunCtx
	Policy string
}

var pod2podSrvTemplate = template.Must(template.New("srv").Parse(`apiVersion: v1
kind: Pod
metadata:
  name: kubenetbench-{{.runID}}-srv
  labels : {
    kubenetbench-runid: {{.runID}},
    role: srv,
  }
spec:
  {{.srvAffinity}}
  containers:
  - {{.srvContainer}}
`))

func (s *Pod2PodSt) genSrvYaml() (string, error) {
	vals := map[string]interface{}{
		"runID":        s.Runctx.id,
		"srvContainer": "{{template \"netperfContainer\"}}",
		"srvAffinity":  "{{template \"srvAffinity\"}}",
	}

	templates := map[string]utils.PrefixRenderer{
		"netperfContainer": s.Runctx.benchmark.WriteSrvContainerYaml,
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
	utils.RenderTemplate(pod2podSrvTemplate, vals, templates, f)
	f.Close()
	return yaml, nil
}

var pod2podPortPolicyTemplate = template.Must(template.New("policy").Parse(`apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: kubenetbench-{{.runID}}-policy
  labels : {
     "kubenetbench-runid": {{.runID}},
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

func (s *Pod2PodSt) genPortPolicyYaml() string {
	m := map[string]interface{}{
		"runID": s.Runctx.id,
	}

	yaml := fmt.Sprintf("%s/port-policy.yaml", s.Runctx.dir)
	if !s.Runctx.quiet {
		log.Printf("Generating %s", yaml)
	}
	f, err := os.Create(yaml)
	if err != nil {
		log.Fatal(err)
	}
	pod2podPortPolicyTemplate.Execute(f, m)
	f.Close()
	return yaml
}

var pod2podCliTemplate = template.Must(template.New("cli").Parse(`apiVersion: v1
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

func (s *Pod2PodSt) genCliYaml(serverIP string) (string, error) {
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

	utils.RenderTemplate(pod2podCliTemplate, vals, templates, f)
	f.Close()
	return yaml, nil
}

// Execute pod2pod command
func (s Pod2PodSt) Execute() error {
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
