package tempcleaner

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

var (
	//go:embed build/*
	buildFS embed.FS

	defaultBinaryPath string
	initErr           error
)

func init() {
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	binName := fmt.Sprintf("temp-cleaner-%s-%s%s", runtime.GOOS, runtime.GOARCH, ext)

	// Unpack from embedded filesystem
	binData, err := buildFS.ReadFile("build/" + binName)
	if err != nil {
		initErr = fmt.Errorf("embedded binary not found for %s/%s: %w", runtime.GOOS, runtime.GOARCH, err)
		return
	}

	// Write it to the temporary directory
	tmpPath := filepath.Join(os.TempDir(), binName)
	if _, err := os.Stat(tmpPath); os.IsNotExist(err) {
		if err := os.WriteFile(tmpPath, binData, 0755); err != nil {
			initErr = fmt.Errorf("failed to write binary to temp dir: %w", err)
			return
		}
	}

	defaultBinaryPath = tmpPath
}

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
		if initErr != nil {
			return 0, fmt.Errorf("default binary initialization failed: %w", initErr)
		}
		if defaultBinaryPath == "" {
			return 0, fmt.Errorf("BinaryPath is required and default embedded binary is not available")
		}
		opts.BinaryPath = defaultBinaryPath
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
