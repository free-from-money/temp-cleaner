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
	cmd := exec.CommandContext(ctx, "lsof", "+D", path, "-t")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()

	// Check stdout before the exit code: lsof exits 1 both when it found nothing
	// and when it found PIDs but failed to walk part of the tree (-t suppresses
	// the warning, so stderr is empty too). Any PID on stdout means in use.
	if stdout.Len() > 0 {
		return true, nil
	}

	if err != nil {
		// Exit 1 with no output: no open files found.
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return false, nil
		}
		// lsof missing or another failure: report it so the caller retries
		// instead of treating the directory as free.
		return false, fmt.Errorf("failed to run lsof: %w", err)
	}

	return false, nil
}
