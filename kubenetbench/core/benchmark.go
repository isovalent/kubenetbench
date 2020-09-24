package core

import (
	"github.com/cilium/kubenetbench/utils"
)

// Benchmark interface
type Benchmark interface {
	// write container server YAML
	WriteSrvContainerYaml(pw *utils.PrefixWriter, params map[string]interface{})
	// write container client YAML
	WriteCliContainerYaml(pw *utils.PrefixWriter, params map[string]interface{})
	// write server ports section
	WriteSrvPortsYaml(pw *utils.PrefixWriter, params map[string]interface{})

	GetTimeout() int
}
