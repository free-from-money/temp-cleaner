//go:build windows

package checker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// osIsInUse checks if a directory is in use on Windows.
// A common way to check if a directory is in use on Windows is to try renaming it
// to its own name or a temporary name.
func osIsInUse(ctx context.Context, path string) (bool, error) {
	// Attempt to rename the directory to a temporary name and back.
	// If a file inside is open, Windows will typically block the directory rename with an access denied error.
	
	// A less invasive check is to try opening the directory with exclusive access,
	// but directories in Windows don't behave exactly like files for exclusive locking.
	// Let's use the rename trick, but rename it to a temp name in the same parent.
	
	parent := filepath.Dir(path)
	base := filepath.Base(path)
	tempName := filepath.Join(parent, base+"_check_tmp")

	err := os.Rename(path, tempName)
	if err != nil {
		// If we get an error, it might be access denied due to being in use.
		// os.IsPermission or check for specific syscall errors.
		// For simplicity, any error during rename (when it exists) we treat as in use.
		if os.IsNotExist(err) {
			return false, nil
		}
		return true, nil
	}

	// Rename it back immediately
	if err := os.Rename(tempName, path); err != nil {
		return false, fmt.Errorf("directory was renamed but failed to rename back: %w", err)
	}

	return false, nil
}
