package core

import (
	"../utils"
)

type Benchmark interface {
	// write container server YAML
	WriteSrvYaml(pw *utils.PrefixWriter, params map[string]interface{})
	// write container client YAML
	WriteCliYaml(pw *utils.PrefixWriter, params map[string]interface{})

	GetTimeout() int
}
