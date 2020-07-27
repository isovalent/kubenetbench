package cmd

import (
	"github.com/spf13/cobra"

	"../core"
)

var quiet bool
var runID string
var runDirBase string
var noCleanup bool

var rootCmd = &cobra.Command{
	Use:   "kubenetbench",
	Short: "kubenetbench is a k8s network benchmark",
}

func init() {

	rootCmd.PersistentFlags().StringVarP(&runID, "runid", "r", "", "run id")
	rootCmd.MarkPersistentFlagRequired("runid")

	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet output")
	rootCmd.PersistentFlags().BoolVar(&noCleanup, "no-cleanup", false, "do not perform cleanup (delete created k8s resources, etc.)")
	rootCmd.PersistentFlags().StringVarP(&runDirBase, "rundir", "d", ".", "base directory to store configuration files and results")

	rootCmd.AddCommand(intrapodCmd)
	rootCmd.AddCommand(serviceCmd)
}

func getRunCtx() *core.RunCtx {
	return core.NewRunCtx(runID, runDirBase, quiet, !noCleanup)
}

// Execute runs the main (root) command
func Execute() error {
	return rootCmd.Execute()
}
