package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"../core"
)

var policyArg string

var pod2podCmd = &cobra.Command{
	Use:   "pod2pod",
	Short: "pod-to-pod network benchmark",
	Run: func(cmd *cobra.Command, args []string) {
		if policyArg != "" && policyArg != "port" {
			log.Fatal("invalid policy: ", policyArg)
		}

		runctx, err := getRunCtx(true)
		if err != nil {
			log.Fatal("initializing run context failed:", err)
		}
		st := core.Pod2PodSt{
			Runctx: runctx,
			Policy: policyArg,
		}
		err = st.Execute()
		if err != nil {
			log.Fatal("pod2pod execution failed:", err)
		}
	},
}

func init() {
	pod2podCmd.Flags().StringVar(&policyArg, "policy", "", "isolation policy (empty or \"port\")")
}
