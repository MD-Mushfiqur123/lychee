//go:build !darwin

package runner

import "fmt"

// Execute is a stub for non-darwin platforms.
func Execute(args []string) error {
	return fmt.Errorf("runner engine is not supported on this platform (requires darwin/MLX)")
}
