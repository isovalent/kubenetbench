package core

import (
	"github.com/cilium/kubenetbench/utils"
)

func (s *ContainerSpec) hostOptsWrite(pw *utils.PrefixWriter, params map[string]interface{}) {
	l := func(s string) {
		pw.AppendNewLineOrDie(s)
	}

	if s.HostNetwork {
		l(`hostNetwork: true`)
	}

	if s.HostIPC {
		l(`hostIPC: true`)
	}

	if s.HostPID {
		l(`hostPID: true`)
	}
}
