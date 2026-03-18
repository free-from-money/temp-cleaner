package tempcleaner

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// CleanOptions holds the configuration for launching the cleaner binary.
type CleanOptions struct {
	// BinaryPath is the path to the compiled temp-cleaner executable.
	BinaryPath string
	// TargetDir is the directory to monitor and delete.
	TargetDir string
	// Interval is the polling frequency to check if the directory is in use (default: 5s).
	Interval time.Duration
	// Timeout is the maximum duration to wait before giving up (0 means no timeout).
	Timeout time.Duration
}

// StartDetached runs the specified temp-cleaner binary as an independent background process.
// It returns the process ID of the launched cleaner, or an error if it failed to start.
// The spawned process will survive even if the parent process exits.
func StartDetached(opts CleanOptions) (int, error) {
	if opts.BinaryPath == "" {
		// Auto-resolve binary path based on runtime OS and Arch
		ext := ""
		if runtime.GOOS == "windows" {
			ext = ".exe"
		}
		opts.BinaryPath = filepath.Join("build", fmt.Sprintf("temp-cleaner-%s-%s%s", runtime.GOOS, runtime.GOARCH, ext))
	}
	if opts.TargetDir == "" {
		return 0, fmt.Errorf("TargetDir is required")
	}
	if opts.Interval <= 0 {
		opts.Interval = 5 * time.Second
	}
	// Timeout of 0 means no timeout, so we just let it be.

	cmd := exec.Command(
		opts.BinaryPath,
		"-target", opts.TargetDir,
		"-interval", opts.Interval.String(),
		"-timeout", opts.Timeout.String(),
	)

	// Set OS-specific detachment attributes
	setDetachedSysProcAttr(cmd)

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start cleaner process: %w", err)
	}

	pid := cmd.Process.Pid

	// Release the process from the Go runtime so it is completely independent.
	// This prevents zombie processes on Unix-like systems and detaches management.
	if err := cmd.Process.Release(); err != nil {
		return pid, fmt.Errorf("process started but failed to release: %w", err)
	}

	return pid, nil
}
