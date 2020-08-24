package core

import (
	"fmt"
	"log"
	"os"
	"text/template"
	"time"

	"../utils"
)

// RunCtx is the context for a benchmark run
type RunCtx struct {
	id          string    // id identifies the run
	dir         string    // directory to store results/etc.
	cliAffinity string    // client affinity
	srvAffinity string    // server affinity
	quiet       bool      // supress output
	cleanup     bool      // perform cleanup: remove k8s entitites (pods, policies, etc.)
	benchmark   Benchmark // underlying benchmark interface
}

// NewRunCtx creates a new RunCtx
func NewRunCtx(
	rid string,
	ridDirBase string,
	cliAffinity string,
	srvAffinity string,
	quiet bool,
	cleanup bool,
	benchmark Benchmark,
) *RunCtx {
	datestr := time.Now().Format("20060102-150405")
	rundir := fmt.Sprintf("%s/%s-%s", ridDirBase, rid, datestr)
	return &RunCtx{
		id:          rid,
		dir:         rundir,
		cliAffinity: cliAffinity,
		srvAffinity: srvAffinity,
		quiet:       quiet,
		cleanup:     cleanup,
		benchmark:   benchmark,
	}
}

func (r *RunCtx) MakeDir() error {
	return os.Mkdir(r.dir, 0755)
}

var runctxCliTemplate = template.Must(template.New("cli").Parse(`apiVersion: v1
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

func (r *RunCtx) genCliYaml(serverIP string) (string, error) {
	yaml := fmt.Sprintf("%s/client.yaml", r.dir)
	if !r.quiet {
		log.Printf("Generating %s", yaml)
	}
	f, err := os.Create(yaml)
	if err != nil {
		return "", err
	}

	vals := map[string]interface{}{
		"runID":        r.id,
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
