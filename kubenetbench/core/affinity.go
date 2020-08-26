package core

import (
	"fmt"
	"strings"

	"github.com/kkourt/kubenetbench/utils"
)

// client on the same node as the server
func cliAffinitySame(pw *utils.PrefixWriter) {
	l := func(s string) {
		pw.AppendNewLineOrDie(s)
	}

	l(`affinity:`)
	l(`   podAffinity:`)
	l(`       requiredDuringSchedulingIgnoredDuringExecution:`)
	l(`       - labelSelector:`)
	l(`            matchExpressions:`)
	l(`            - key: role`)
	l(`              operator: In`)
	l(`              values:`)
	l(`              - srv`)
	l(`         topologyKey: "kubernetes.io/hostname"`)
}

// client on the same node as the server
func cliAffinityOther(pw *utils.PrefixWriter) {
	l := func(s string) {
		pw.AppendNewLineOrDie(s)
	}

	l(`affinity:`)
	l(`   podAntiAffinity:`)
	l(`       requiredDuringSchedulingIgnoredDuringExecution:`)
	l(`       - labelSelector:`)
	l(`            matchExpressions:`)
	l(`            - key: role`)
	l(`              operator: In`)
	l(`              values:`)
	l(`              - srv`)
	l(`         topologyKey: "kubernetes.io/hostname"`)
}

//
func affinityHost(host string, pw *utils.PrefixWriter) {
	pw.AppendNewLineOrDie(`nodeSelector:`)
	pw.AppendNewLineOrDie(fmt.Sprintf(`     kubernetes.io/hostname: %s`, host))
}

func (c *RunBenchCtx) cliAffinityWrite(pw *utils.PrefixWriter, params map[string]interface{}) {
	switch {
	case c.cliAffinity == "none":
		return
	case c.cliAffinity == "same":
		cliAffinitySame(pw)
	case c.cliAffinity == "different":
		cliAffinityOther(pw)
	case strings.HasPrefix(c.cliAffinity, "host="):
		host := strings.TrimPrefix(c.cliAffinity, "host=")
		affinityHost(host, pw)

	default:
		panic(fmt.Sprintf("Unrecognized client affinity: %s", c.cliAffinity))
	}
}

func (c *RunBenchCtx) srvAffinityWrite(pw *utils.PrefixWriter, params map[string]interface{}) {
	switch {
	case c.srvAffinity == "none":
		return
	case strings.HasPrefix(c.srvAffinity, "host="):
		host := strings.TrimPrefix(c.srvAffinity, "host=")
		affinityHost(host, pw)

	default:
		panic(fmt.Sprintf("Unrecognized server affinity: %s", c.srvAffinity))
	}
}
