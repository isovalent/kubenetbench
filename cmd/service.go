package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"../core"
)

var serviceTypeArg string

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "service network benchmark",
	Run: func(cmd *cobra.Command, args []string) {

		if serviceTypeArg != "ClusterIP" {
			log.Fatal("invalid policy: ", serviceTypeArg)
		}

		st := core.ServiceSt{
			Runctx:      getRunCtx(),
			NetperfConf: getNetperfConf(),
			ServiceType: serviceTypeArg,
		}
		st.Execute()
	},
}

func init() {
	serviceCmd.Flags().StringVar(&serviceTypeArg, "type", "ClusterIP", "service type (ClusterIP)")
}
