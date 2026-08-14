//go:build !windows

package checker

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func requireLsof(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("lsof"); err != nil {
		t.Skip("lsof not available")
	}
}

func TestIsInUseWithOpenFile(t *testing.T) {
	requireLsof(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	inUse, err := New().IsInUse(context.Background(), dir)
	if err != nil {
		t.Fatalf("IsInUse: %v", err)
	}
	if !inUse {
		t.Errorf("IsInUse(%s) = false, want true (this process holds %s open)", dir, path)
	}
}

func TestIsInUseIdle(t *testing.T) {
	requireLsof(t)

	dir := t.TempDir()

	inUse, err := New().IsInUse(context.Background(), dir)
	if err != nil {
		t.Fatalf("IsInUse: %v", err)
	}
	if inUse {
		t.Errorf("IsInUse(%s) = true, want false (nothing open)", dir)
	}
}
