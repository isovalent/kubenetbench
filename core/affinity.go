package core

import (
	"fmt"

	"../utils"
)

// client on the same node as the server
func cliAffinitySame(pw *utils.PrefixWriter, params map[string]interface{}) {
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
func cliAffinityOther(pw *utils.PrefixWriter, params map[string]interface{}) {
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

// no afinity
func (c *RunCtx) cliAffinityWrite(pw *utils.PrefixWriter, params map[string]interface{}) {
	switch c.cliAffinity {
	case "none":
		return
	case "same":
		cliAffinitySame(pw, params)
	case "other":
		cliAffinityOther(pw, params)
	default:
		panic(fmt.Sprintf("Unrecognized affinity: %s", c.cliAffinity))
	}
}
