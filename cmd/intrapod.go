package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"../core"
)

var policyArg string

var intrapodCmd = &cobra.Command{
	Use:   "intrapod",
	Short: "pod-to-pod network benchmark",
	Run: func(cmd *cobra.Command, args []string) {
		if policyArg != "" && policyArg != "port" {
			log.Fatal("invalid policy: ", policyArg)
		}

		st := core.IntrapodSt{
			Runctx:      getRunCtx(),
			NetperfConf: getNetperfConf(),
			Policy:      policyArg,
		}
		err := st.Execute()
		if err != nil {
			log.Fatal("intrapod execution failed:", err)
		}
	},
}

func init() {
	intrapodCmd.Flags().StringVar(&policyArg, "policy", "", "isolation policy (empty or \"port\")")
}
