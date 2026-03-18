package deleter

import (
	"context"
	"fmt"
	"os"
	"time"
)

// Deleter defines the interface for deleting directories.
// This interface is defined where it's conceptually used (in the main package's logic),
// but since main shouldn't be imported, we define a concrete type here and let main use it directly.
type Deleter struct {
	MaxRetries int
	RetryDelay time.Duration
}

// New creates a new Deleter instance.
func New() *Deleter {
	return &Deleter{
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}
}

// RemoveAll attempts to remove a directory and its contents, with retries.
func (d *Deleter) RemoveAll(ctx context.Context, path string) error {
	var lastErr error
	for i := 0; i <= d.MaxRetries; i++ {
		// Check context cancellation
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context cancelled during deletion: %w", err)
		}

		err := os.RemoveAll(path)
		if err == nil {
			return nil
		}
		lastErr = err

		if i < d.MaxRetries {
			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during deletion retry wait: %w", ctx.Err())
			case <-time.After(d.RetryDelay):
				// Wait and retry
			}
		}
	}
	return fmt.Errorf("failed to remove %q after %d retries: %w", path, d.MaxRetries, lastErr)
}
