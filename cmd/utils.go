package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"
)

// ExecCmd: execute a command using the shell
func (c *RunCtx) ExecCmd(argcmd string) {
	if !c.quiet {
		log.Println("$", argcmd)
	}
	ctx, cancel := context.WithTimeout(context.Background(), CmdTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", []string{"-c", argcmd}...)
	err := cmd.Run()
	if err != nil {
		log.Fatal("Command '", argcmd, "' failed: ", err)
	}
}

func execCmdLines(argcmd string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), CmdTimeout)
	defer cancel()
	var ret []string

	cmd := exec.CommandContext(ctx, "sh", []string{"-c", argcmd}...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ret, err
	}

	err = cmd.Start()
	if err != nil {
		return ret, err
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		ret = append(ret, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return ret, err
	}

	err = cmd.Wait()
	return ret, err
}

func (c *RunCtx) ExecCmdLines(argcmd string) []string {

	if !c.quiet {
		log.Println("$", argcmd)
	}
	lines, err := execCmdLines(argcmd)
	if err != nil {
		log.Fatal("Error executing:", argcmd, " :", err)
	}
	return lines
}

func (c *RunCtx) KubeGetIP(
	selector string,
	retries uint,
	sleept time.Duration,
) string {

	cmd := fmt.Sprintf(
		"kubectl get pod -l \"%s\" -o custom-columns=IP:.status.podIP --no-headers",
		selector,
	)
	for {
		if !c.quiet {
			log.Printf("$ %s # (remaining retries: %d)", cmd, retries)
		}
		lines, err := execCmdLines(cmd)
		if err == nil && len(lines) == 1 && lines[0] != "<none>" {
			return lines[0]
		}

		if retries == 0 {
			log.Fatal("Error executing:", cmd, " :", err)
		}

		retries--
		time.Sleep(sleept)
	}
}

func (c *RunCtx) KubeGetPhase(selector string) string {
	cmd := fmt.Sprintf(
		"kubectl get pod -l \"%s\" -o custom-columns=Status:.status.phase --no-headers",
		selector,
	)

	lines, err := execCmdLines(cmd)
	if err != nil {
		log.Fatal("Error executing:", cmd, " :", err)
	}

	if len(lines) != 1 {
		log.Fatal("Selector did not provide single result cmd:", cmd, "result:", lines)
	}

	return lines[0]
}

func (c *RunCtx) KubeSaveLogs(selector string, logfile string) {
	argcmd := fmt.Sprintf("kubectl logs -l \"%s\" > %s", selector, logfile)
	c.ExecCmd(argcmd)
}
