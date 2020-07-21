package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"time"
)

type RunCtx struct {
	id    string
	dir   string
	quiet bool
}

var quiet bool
var runId string
var runDirBase string

var rootCmd = &cobra.Command{
	Use:   "kubenetbench",
	Short: "kubenetbench is a k8s network benchmark",
}

const CmdTimeout = 90 * time.Second

func init() {

	rootCmd.PersistentFlags().StringVarP(&runId, "runid", "r", "", "run id")
	rootCmd.MarkPersistentFlagRequired("runid")

	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet output")
	rootCmd.PersistentFlags().StringVarP(&runDirBase, "rundir", "d", ".", "base directory to store configuration files and results")

	rootCmd.AddCommand(intrapodCmd)
}

func Execute() error {
	return rootCmd.Execute()
}

func newRunCtx() *RunCtx {
	datestr := time.Now().Format("20060102-150405")
	rundir := fmt.Sprintf("%s/%s-%s", runDirBase, runId, datestr)
	err := os.Mkdir(rundir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	return &RunCtx{
		id:    runId,
		dir:   rundir,
		quiet: quiet,
	}
}
