//go:build !windows

package main

import (
	"os"
	"os/exec"
)

func prepareInterruptible(cmd *exec.Cmd) {}

func interrupt(cmd *exec.Cmd) error { return cmd.Process.Signal(os.Interrupt) }
