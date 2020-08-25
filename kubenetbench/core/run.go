package core

import (
	"fmt"
	"log"
	"os"
	"text/template"
	"time"

	"github.com/kkourt/kubenetbench/utils"
)

// RunBenchCtx is the context for a benchmark run
type RunBenchCtx struct {
	session     *Session  // session
	runid       string    //
	cliAffinity string    // client affinity
	srvAffinity string    // server affinity
	cleanup     bool      // perform cleanup: remove k8s entitites (pods, policies, etc.)
	benchmark   Benchmark // underlying benchmark interface
}

func NewRunBenchCtx(
	sess *Session,
	runLabel string,
	cliAffinity string,
	srvAffinity string,
	cleanup bool,
	benchmark Benchmark,
) *RunBenchCtx {
	datestr := time.Now().Format("20060102-150405")
	runid := fmt.Sprintf("%s-%s", runLabel, datestr)
	return &RunBenchCtx{
		session:     sess,
		runid:       runid,
		cliAffinity: cliAffinity,
		srvAffinity: srvAffinity,
		cleanup:     cleanup,
		benchmark:   benchmark,
	}
}

func (r *RunBenchCtx) getRunLabel(sep string) string {
	return fmt.Sprintf("%s%s%s", runIdLabel, sep, r.runid)
}

func (r *RunBenchCtx) getDir() string {
	return fmt.Sprintf("%s/%s", r.session.dir, r.runid)
}

func (r *RunBenchCtx) MakeDir() error {
	d := r.getDir()
	return os.Mkdir(d, 0755)
}

var runctxCliTemplate = template.Must(template.New("cli").Parse(`apiVersion: v1
kind: Pod
metadata:
  name: knb-cli
  labels : {
     {{.runLabel}},
     role: cli,
  }
spec:
  restartPolicy: Never
  {{.cliAffinity}}
  containers:
  - {{.cliContainer}}
`))

func (r *RunBenchCtx) genCliYaml(serverIP string) (string, error) {
	yaml := fmt.Sprintf("%s/client.yaml", r.getDir())
	log.Printf("Generating %s", yaml)
	f, err := os.Create(yaml)
	if err != nil {
		return "", err
	}

	vals := map[string]interface{}{
		"runLabel":     r.getRunLabel(": "),
		"serverIP":     serverIP,
		"cliContainer": "{{template \"netperfContainer\"}}",
		"cliAffinity":  "{{template \"cliAffinity\"}}",
	}

	templates := map[string]utils.PrefixRenderer{
		"netperfContainer": r.benchmark.WriteCliContainerYaml,
		"cliAffinity":      r.cliAffinityWrite,
	}

	utils.RenderTemplate(runctxCliTemplate, vals, templates, f)
	f.Close()
	return yaml, nil
}
