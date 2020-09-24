package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/cilium/kubenetbench/kubenetbench/core"
)

var serviceTypeArg string

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "service network benchmark run",
	Run: func(cmd *cobra.Command, args []string) {

		if serviceTypeArg != "ClusterIP" {
			log.Fatal("invalid policy: ", serviceTypeArg)
		}

		runctx, err := getRunBenchCtx(serviceTypeArg, true)
		if err != nil {
			log.Fatal("initializing run context failed:", err)
		}
		st := core.ServiceSt{
			RunBenchCtx: runctx,
			ServiceType: serviceTypeArg,
		}
		err = st.Execute()
		if err != nil {
			log.Fatal("service execution failed:", err)
		}
	},
}

func init() {
	addBenchmarkFlags(serviceCmd)
	serviceCmd.Flags().StringVar(&serviceTypeArg, "type", "ClusterIP", "service type (ClusterIP)")
}
