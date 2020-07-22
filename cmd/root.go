package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"time"
)

type runCtx struct {
	id      string
	dir     string
	quiet   bool
	cleanup bool
}

var quiet bool
var runID string
var runDirBase string
var noCleanup bool

var rootCmd = &cobra.Command{
	Use:   "kubenetbench",
	Short: "kubenetbench is a k8s network benchmark",
}

const cmdTimeout = 90 * time.Second

func init() {

	rootCmd.PersistentFlags().StringVarP(&runID, "runid", "r", "", "run id")
	rootCmd.MarkPersistentFlagRequired("runid")

	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet output")
	rootCmd.PersistentFlags().BoolVar(&noCleanup, "no-cleanup", false, "do not perform cleanup (delete created k8s resources, etc.)")
	rootCmd.PersistentFlags().StringVarP(&runDirBase, "rundir", "d", ".", "base directory to store configuration files and results")

	rootCmd.AddCommand(intrapodCmd)
	rootCmd.AddCommand(serviceCmd)
}

// Execute runs the main (root) command
func Execute() error {
	return rootCmd.Execute()
}

func newRunCtx() *runCtx {
	datestr := time.Now().Format("20060102-150405")
	rundir := fmt.Sprintf("%s/%s-%s", runDirBase, runID, datestr)
	err := os.Mkdir(rundir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	return &runCtx{
		id:      runID,
		dir:     rundir,
		quiet:   quiet,
		cleanup: !noCleanup,
	}
}
