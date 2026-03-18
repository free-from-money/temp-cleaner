//go:build !windows

package checker

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// osIsInUse checks if a directory is in use on Unix-like systems using lsof.
func osIsInUse(ctx context.Context, path string) (bool, error) {
	// lsof +D /path/to/dir checks for open files in the directory.
	// We use the -t flag to only get PIDs, making it terser.
	// If lsof finds open files, it exits with 0. If it finds none, it typically exits with 1.
	cmd := exec.CommandContext(ctx, "lsof", "+D", path, "-t")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		// lsof returns exit status 1 if no files are open (or if it's not found/errors).
		// We should differentiate between "no files open" and actual errors.
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				return false, nil // No files are open
			}
		}
		// In case lsof is not installed or another error occurs.
		return false, fmt.Errorf("failed to run lsof: %w", err)
	}

	// If it ran successfully and output something, it means files are open.
	if stdout.Len() > 0 {
		return true, nil
	}

	return false, nil
}
