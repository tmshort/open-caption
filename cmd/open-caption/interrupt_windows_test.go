//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

// Windows has no signals: the way to interrupt a console program is a Ctrl+Break event sent to its process
// group, so the child is started in a group of its own.
func prepareInterruptible(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}

func interrupt(cmd *exec.Cmd) error {
	d, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return err
	}
	p, err := d.FindProc("GenerateConsoleCtrlEvent")
	if err != nil {
		return err
	}
	if r, _, err := p.Call(uintptr(syscall.CTRL_BREAK_EVENT), uintptr(cmd.Process.Pid)); r == 0 {
		return err
	}
	return nil
}
