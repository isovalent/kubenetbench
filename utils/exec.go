package utils

import (
	"bufio"
	"context"
	"os/exec"
	"time"
)

// global command timeout parameter
const cmdTimeout = 90 * time.Second

// ExecCmd executes a command using the shell
func ExecCmd(argcmd string) error {
	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", []string{"-c", argcmd}...)
	err := cmd.Run()
	return err
}

// ExecCmdLines executes a command using the shell, and return the output lines or an error
func ExecCmdLines(argcmd string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
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
