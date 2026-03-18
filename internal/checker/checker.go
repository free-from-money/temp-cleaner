package checker

import "context"

// Checker provides functionality to check if a directory is currently in use.
type Checker struct{}

// New creates a new Checker.
func New() *Checker {
	return &Checker{}
}

// IsInUse checks if the target path is currently being used (locked) by any process.
func (c *Checker) IsInUse(ctx context.Context, path string) (bool, error) {
	return osIsInUse(ctx, path)
}
