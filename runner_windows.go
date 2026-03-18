//go:build windows

package tempcleaner

import (
	"os/exec"
	"syscall"
)

// setDetachedSysProcAttr configures the command to run as a detached process on Windows.
func setDetachedSysProcAttr(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	// DETACHED_PROCESS (0x00000008) flag in Windows creates a new process 
	// without attached console, detaching it from the parent process.
	cmd.SysProcAttr.CreationFlags = 0x00000008
}
