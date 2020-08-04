package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"../core"
)

var quiet bool
var runID string
var runDirBase string
var noCleanup bool
var benchmark string
var benchmarkDuration int
var cliAffinity string

var rootCmd = &cobra.Command{
	Use:   "kubenetbench",
	Short: "kubenetbench is a k8s network benchmark",
}

var nopCmd = &cobra.Command{
	Use:   "nop",
	Short: "does nothing (used for testing)",
	Run:   func(cmd *cobra.Command, args []string) {},
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "clean k8s resources",
	Run: func(cmd *cobra.Command, args []string) {
		runctx, err := getRunCtx(false)
		if err != nil {
			log.Fatal("initializing run context failed:", err)
		}
		runctx.KubeCleanup()
	},
}

func init() {

	rootCmd.PersistentFlags().StringVarP(&runID, "runid", "i", "", "run id")
	rootCmd.MarkPersistentFlagRequired("runid")

	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet output")
	rootCmd.PersistentFlags().BoolVar(&noCleanup, "no-cleanup", false, "do not perform cleanup (delete created k8s resources, etc.)")
	rootCmd.PersistentFlags().StringVarP(&runDirBase, "rundir", "r", ".", "base directory to store configuration files and results")
	rootCmd.PersistentFlags().StringVarP(&benchmark, "benchmark", "b", "netperf", "benchmark to use")
	rootCmd.PersistentFlags().IntVarP(&benchmarkDuration, "duration", "d", 30, "benchmark duration (sec)")
	rootCmd.PersistentFlags().StringVar(&cliAffinity, "client-affinity", "none", "client affinity (none, same: same as server, differnt: different than server)")

	rootCmd.AddCommand(nopCmd)
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(intrapodCmd)
	rootCmd.AddCommand(serviceCmd)
}

func getRunCtx(mkdir bool) (*core.RunCtx, error) {

	switch benchmark {
	case "netperf":
		netperfBench := getNetperfBench()
		ctx := core.NewRunCtx(
			runID,
			runDirBase,
			cliAffinity,
			quiet,
			!noCleanup,
			netperfBench)

		var err error = nil
		if mkdir {
			err = ctx.MakeDir()
			if err != nil {
				ctx = nil
			}
		}

		return ctx, err

	case "ipperf":
		return nil, fmt.Errorf("NYI: %s", benchmark)

	}

	return nil, fmt.Errorf("unknown benchmark: %s", benchmark)
}

// Execute runs the main (root) command
func Execute() error {
	return rootCmd.Execute()
}
