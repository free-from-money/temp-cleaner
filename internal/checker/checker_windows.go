//go:build windows

package checker

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

// osIsInUse checks whether anything under path is held open by another process.
//
// Every entry in the tree is opened with dwShareMode=0, which asks Windows for
// exclusive access. If some other process already holds the entry open, the
// open fails with ERROR_SHARING_VIOLATION, and that is exactly the condition
// that would make deleting it fail too. Nothing under path is modified.
func osIsInUse(ctx context.Context, path string) (bool, error) {
	inUse := false

	err := filepath.WalkDir(path, func(name string, _ fs.DirEntry, walkErr error) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if walkErr != nil {
			// Gone (including the top-level path never existing): nothing to hold.
			if os.IsNotExist(walkErr) {
				return nil
			}
			// Cannot even list it, so we cannot delete it.
			inUse = true
			return filepath.SkipAll
		}
		if locked(name) {
			inUse = true
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return inUse, nil
}

// locked reports whether name cannot be opened exclusively.
// FILE_FLAG_BACKUP_SEMANTICS is required to get a handle to a directory at all,
// and is what catches a process holding the directory as its working directory.
func locked(name string) bool {
	p, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return true
	}
	h, err := syscall.CreateFile(p, syscall.GENERIC_READ, 0, nil,
		syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		// Vanished mid-walk: not in use. Anything else (sharing violation,
		// access denied) means we could not delete it either, so say in use.
		return !errors.Is(err, syscall.ERROR_FILE_NOT_FOUND) &&
			!errors.Is(err, syscall.ERROR_PATH_NOT_FOUND)
	}
	syscall.CloseHandle(h)
	return false
}
