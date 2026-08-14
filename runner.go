package tempcleaner

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// unpackBinary decompresses the embedded cleaner binary into the temp directory.
// It runs at most once per process, and only when StartDetached actually needs it.
// The file name carries a hash of the embedded data, so a new library version
// never reuses a stale binary left behind in the temp directory.
var unpackBinary = sync.OnceValues(func() (string, error) {
	if len(binGz) == 0 {
		return "", fmt.Errorf("embedded binary not found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	sum := sha256.Sum256(binGz)
	binName := fmt.Sprintf("temp-cleaner-%s-%s-%s%s", runtime.GOOS, runtime.GOARCH, hex.EncodeToString(sum[:4]), ext)
	tmpPath := filepath.Join(os.TempDir(), binName)

	if _, err := os.Stat(tmpPath); err == nil {
		return tmpPath, nil
	}

	zr, err := gzip.NewReader(bytes.NewReader(binGz))
	if err != nil {
		return "", fmt.Errorf("failed to decompress embedded binary: %w", err)
	}
	defer zr.Close()

	binData, err := io.ReadAll(zr)
	if err != nil {
		return "", fmt.Errorf("failed to decompress embedded binary: %w", err)
	}

	// Write to a sibling temp file and rename, so a concurrent process never
	// execs a half-written binary.
	f, err := os.CreateTemp(filepath.Dir(tmpPath), binName+".*")
	if err != nil {
		return "", fmt.Errorf("failed to write binary to temp dir: %w", err)
	}
	defer os.Remove(f.Name())

	if _, err := f.Write(binData); err != nil {
		f.Close()
		return "", fmt.Errorf("failed to write binary to temp dir: %w", err)
	}
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("failed to write binary to temp dir: %w", err)
	}
	if err := os.Chmod(f.Name(), 0755); err != nil {
		return "", fmt.Errorf("failed to write binary to temp dir: %w", err)
	}
	if err := os.Rename(f.Name(), tmpPath); err != nil {
		return "", fmt.Errorf("failed to write binary to temp dir: %w", err)
	}

	return tmpPath, nil
})

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
		path, err := unpackBinary()
		if err != nil {
			return 0, fmt.Errorf("default binary initialization failed: %w", err)
		}
		opts.BinaryPath = path
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
