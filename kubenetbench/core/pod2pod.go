package core

import (
	"fmt"
	"log"
	"os"
	"text/template"
	"time"

	"github.com/cilium/kubenetbench/utils"
)

// Pod2PodSt is the necessary state for executing an pod-to-pod benchmark
type Pod2PodSt struct {
	RunBenchCtx *RunBenchCtx
	Policy      string
}

var pod2podSrvTemplate = template.Must(template.New("srv").Parse(`apiVersion: v1
kind: Pod
metadata:
  name: knb-srv
  labels : {
    {{.sessLabel}},
    {{.runLabel}},
    role: srv,
  }
spec:
  {{.srvSpec}}
  containers:
  - {{.srvContainer}}
`))

func (s *Pod2PodSt) genSrvYaml() (string, error) {
	vals := map[string]interface{}{
		"sessLabel":    s.RunBenchCtx.session.getSessionLabel(": "),
		"runLabel":     s.RunBenchCtx.getRunLabel(": "),
		"srvContainer": "{{template \"netperfContainer\"}}",
		"srvSpec":      "{{template \"srvSpec\"}}",
	}

	templates := map[string]utils.PrefixRenderer{
		"netperfContainer": s.RunBenchCtx.benchmark.WriteSrvContainerYaml,
		"srvSpec":          s.RunBenchCtx.srvPodSpecWrite,
	}

	yaml := fmt.Sprintf("%s/netserv.yaml", s.RunBenchCtx.getDir())
	log.Printf("Generating %s", yaml)
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
		"runID": s.RunBenchCtx.runid,
	}

	yaml := fmt.Sprintf("%s/port-policy.yaml", s.RunBenchCtx.getDir())
	log.Printf("Generating %s", yaml)
	f, err := os.Create(yaml)
	if err != nil {
		log.Fatal(err)
	}
	pod2podPortPolicyTemplate.Execute(f, m)
	f.Close()
	return yaml
}

func (s *Pod2PodSt) genCliYaml(serverIP string) (string, error) {
	return s.RunBenchCtx.genCliYaml(serverIP)
}

// Execute pod2pod command
func (s Pod2PodSt) Execute() error {
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

	// get server pod IP
	time.Sleep(2 * time.Second)
	srvIP, err := s.RunBenchCtx.KubeGetPodIP(srvSelector, 30, 2*time.Second)
	if err != nil {
		return err
	}
	log.Printf("server_ip=%s", srvIP)

	// start policy if specified
	if s.Policy == "port" {
		policyYamlFname := s.genPortPolicyYaml()
		err := s.RunBenchCtx.KubeApply(policyYamlFname)
		if err != nil {
			return fmt.Errorf("failed to apply policy: %w", err)
		}
	}

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

	return s.RunBenchCtx.finalizeAndWait()
}
