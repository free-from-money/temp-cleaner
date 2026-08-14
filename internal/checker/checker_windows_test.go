//go:build windows

package checker

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// makeTarget builds <tempdir>/target/nested/a.txt and returns the target dir
// and the file. The extra level means the parent holds nothing but target,
// so leftovers from a destructive probe are easy to spot.
func makeTarget(t *testing.T) (dir, file string) {
	t.Helper()
	dir = filepath.Join(t.TempDir(), "target")
	if err := os.MkdirAll(filepath.Join(dir, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	file = filepath.Join(dir, "nested", "a.txt")
	if err := os.WriteFile(file, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir, file
}

func TestIsInUseWithOpenFile(t *testing.T) {
	dir, file := makeTarget(t)

	f, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	inUse, err := New().IsInUse(context.Background(), dir)
	if err != nil {
		t.Fatalf("IsInUse: %v", err)
	}
	if !inUse {
		t.Errorf("IsInUse(%s) = false, want true (this process holds %s open)", dir, file)
	}
}

func TestIsInUseIdle(t *testing.T) {
	dir, _ := makeTarget(t)

	inUse, err := New().IsInUse(context.Background(), dir)
	if err != nil {
		t.Fatalf("IsInUse: %v", err)
	}
	if inUse {
		t.Errorf("IsInUse(%s) = true, want false (nothing open)", dir)
	}
}

// TestIsInUseIsNonDestructive guards against the old rename-based probe, which
// moved the target to <base>_check_tmp and could leave it there.
func TestIsInUseIsNonDestructive(t *testing.T) {
	dir, file := makeTarget(t)
	parent := filepath.Dir(dir)

	f, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}

	// Once while held open (rename fails), once idle (rename would succeed).
	for _, state := range []string{"open", "closed"} {
		if state == "closed" {
			f.Close()
		}
		if _, err := New().IsInUse(context.Background(), dir); err != nil {
			t.Fatalf("IsInUse (%s): %v", state, err)
		}

		if _, err := os.Stat(dir); err != nil {
			t.Fatalf("target missing after IsInUse (%s): %v", state, err)
		}
		if _, err := os.Stat(file); err != nil {
			t.Fatalf("target file missing after IsInUse (%s): %v", state, err)
		}
		entries, err := os.ReadDir(parent)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 1 || entries[0].Name() != "target" {
			var names []string
			for _, e := range entries {
				names = append(names, e.Name())
			}
			t.Fatalf("IsInUse (%s) left residue in %s: %v, want [target]", state, parent, names)
		}
	}
}
