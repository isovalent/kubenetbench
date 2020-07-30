package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"../core"
)

var quiet bool
var runID string
var runDirBase string
var noCleanup bool
var benchmark string
var benchmarkDuration int

var rootCmd = &cobra.Command{
	Use:   "kubenetbench",
	Short: "kubenetbench is a k8s network benchmark",
}

var nopCmd = &cobra.Command{
	Use:   "nop",
	Short: "does nothing (used for testing)",
	Run:   func(cmd *cobra.Command, args []string) {},
}

func init() {

	rootCmd.PersistentFlags().StringVarP(&runID, "runid", "i", "", "run id")
	rootCmd.MarkPersistentFlagRequired("runid")

	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet output")
	rootCmd.PersistentFlags().BoolVar(&noCleanup, "no-cleanup", false, "do not perform cleanup (delete created k8s resources, etc.)")
	rootCmd.PersistentFlags().StringVarP(&runDirBase, "rundir", "r", ".", "base directory to store configuration files and results")
	rootCmd.PersistentFlags().StringVarP(&benchmark, "benchmark", "b", "netperf", "benchmark to use")
	rootCmd.PersistentFlags().IntVarP(&benchmarkDuration, "duration", "d", 10, "benchmark duration (sec)")

	rootCmd.AddCommand(nopCmd)
	rootCmd.AddCommand(intrapodCmd)
	rootCmd.AddCommand(serviceCmd)
}

func getRunCtx() (*core.RunCtx, error) {

	switch benchmark {
	case "netperf":
		netperfBench := getNetperfBench()
		return core.NewRunCtx(runID, runDirBase, quiet, !noCleanup, netperfBench), nil

	case "ipperf":
		return nil, fmt.Errorf("NYI: %s", benchmark)

	}

	return nil, fmt.Errorf("unknown benchmark: %s", benchmark)
}

// Execute runs the main (root) command
func Execute() error {
	return rootCmd.Execute()
}
