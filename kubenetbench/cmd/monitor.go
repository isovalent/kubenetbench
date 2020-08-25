package cmd

import (
	"github.com/spf13/cobra"
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "benchmark monitor utilities",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var monitorStartCmd = &cobra.Command{
	Use:   "start",
	Short: "start monitor",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var monitorStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "query monitor status",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	monitorCmd.AddCommand(monitorStartCmd)
	monitorCmd.AddCommand(monitorStatusCmd)
}
