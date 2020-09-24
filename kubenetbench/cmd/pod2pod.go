package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/cilium/kubenetbench/kubenetbench/core"
)

var policyArg string

var pod2podCmd = &cobra.Command{
	Use:   "pod2pod",
	Short: "pod-to-pod network benchmark run",
	Run: func(cmd *cobra.Command, args []string) {
		if policyArg != "" && policyArg != "port" {
			log.Fatal("invalid policy: ", policyArg)
		}

		runctx, err := getRunBenchCtx("pod2pod", true)
		if err != nil {
			log.Fatal("initializing run context failed:", err)
		}
		st := core.Pod2PodSt{
			RunBenchCtx: runctx,
			Policy:      policyArg,
		}
		err = st.Execute()
		if err != nil {
			log.Fatal("pod2pod execution failed:", err)
		}
	},
}

func init() {
	addBenchmarkFlags(pod2podCmd)
	pod2podCmd.Flags().StringVar(&policyArg, "policy", "", "isolation policy (empty or \"port\")")
}
