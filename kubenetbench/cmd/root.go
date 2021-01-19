package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cilium/kubenetbench/kubenetbench/core"
)

var (
	quiet           bool
	sessID          string
	sessDirBase     string
	sessPortForward bool
)

// var noCleanup bool

var rootCmd = &cobra.Command{
	Use:   "kubenetbench",
	Short: "kubenetbench is a k8s network benchmark",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "initalize a seasson",
	Run: func(cmd *cobra.Command, args []string) {
		sess, err := core.InitSession(sessID, sessDirBase, sessPortForward)
		if err != nil {
			log.Fatal(fmt.Sprintf("error initializing session: %w", err))
		}
		InitLog(sess)
		log.Printf("Starting session monitor")
		err = sess.StartMonitor()
		if err != nil {
			log.Fatal(fmt.Errorf("failed to start monitor: %w", err))
		}

		err = sess.GetSysInfoNodes()
		if err != nil {
			log.Printf("failed to get (some) sysinfo via monitor: %s", err)
		}
	},
}

var doneCmd = &cobra.Command{
	Use:   "done",
	Short: "terminate the seasson (kill the monitor)",
	Run: func(cmd *cobra.Command, args []string) {
		sess := getSession()
		log.Printf("Starting session monitor")
		err := sess.StopMonitor()
		if err != nil {
			log.Fatal(fmt.Errorf("failed to stop monitor: %w", err))
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&sessID, "session-id", "s", "", "session id")
	rootCmd.MarkPersistentFlagRequired("session-id")
	rootCmd.PersistentFlags().StringVarP(&sessDirBase, "session-base-dir", "d", ".", "base directory to store session data")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet output")
	rootCmd.PersistentFlags().BoolVarP(&sessPortForward, "port-forward", "", false, "use port-forward to connect to monitor")

	// session commands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(doneCmd)

	// benchmark commands
	rootCmd.AddCommand(pod2podCmd)
	rootCmd.AddCommand(serviceCmd)
}

// return a session based on the given flags
func getSession() *core.Session {
	sess, err := core.NewSession(sessID, sessDirBase, sessPortForward)
	if err != nil {
		log.Fatal(fmt.Errorf("error creating session: %w", err))
	}

	InitLog(sess)
	return sess
}

func InitLog(sess *core.Session) {
	f, err := sess.OpenLog()
	if err != nil {
		log.Fatal(fmt.Sprintf("error openning session log file: %w", err))
	}

	if quiet {
		log.SetOutput(f)
	} else {
		m := io.MultiWriter(f, os.Stdout)
		log.SetOutput(m)
	}
	log.Printf("****** %s\n", strings.Join(os.Args, " "))
}

// Execute runs the main (root) command
func Execute() error {
	return rootCmd.Execute()
}
