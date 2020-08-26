package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kkourt/kubenetbench/kubenetbench/core"
)

var (
	benchmark         string
	runLabel          string
	benchmarkDuration int
	cliAffinity       string
	srvAffinity       string
	noCleanup         bool
	collectPerf       bool
)

// add common benchmark flags
func addBenchmarkFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&benchmark, "benchmark", "b", "netperf", "benchmark program to use")
	cmd.Flags().StringVarP(&runLabel, "run-label", "l", "", "benchmark run label")
	cmd.Flags().IntVarP(&benchmarkDuration, "duration", "t", 30, "benchmark duration (sec)")
	cmd.Flags().BoolVar(&noCleanup, "no-cleanup", false, "do not perform cleanup (delete created k8s resources, etc.)")
	cmd.Flags().StringVar(&cliAffinity, "client-affinity", "none", "client affinity (none, same: same as server, different: different than server, host=XXXX)")
	cmd.Flags().StringVar(&srvAffinity, "server-affinity", "none", "server affinity (none, host=XXXX)")
	cmd.Flags().BoolVar(&collectPerf, "collect-perf", false, "collect performance data usning perf")
	addNetperfFlags(cmd)
}

func getRunBenchCtx(defaultRunLabel string, mkdir bool) (*core.RunBenchCtx, error) {
	var bench core.Benchmark

	switch benchmark {
	case "netperf":
		bench = getNetperfBench()
	case "ipperf":
		return nil, fmt.Errorf("benchmark NYI: %s", benchmark)
	default:
		return nil, fmt.Errorf("unknown benchmark: %s", benchmark)
	}

	if runLabel == "" {
		runLabel = defaultRunLabel
	}

	sess := getSession()
	ctx := core.NewRunBenchCtx(
		sess,
		runLabel,
		cliAffinity,
		srvAffinity,
		!noCleanup,
		bench,
		collectPerf)

	var err error = nil
	if mkdir {
		err = ctx.MakeDir()
		if err != nil {
			ctx = nil
		}
	}

	return ctx, err

}
