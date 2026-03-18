//go:build !windows

package tempcleaner

import (
	"os/exec"
	"syscall"
)

// setDetachedSysProcAttr configures the command to run as a detached process on Unix-like systems.
func setDetachedSysProcAttr(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	// Setsid creates a new session and sets the process group ID, 
	// disconnecting the process from its parent's terminal.
	cmd.SysProcAttr.Setsid = true
}
