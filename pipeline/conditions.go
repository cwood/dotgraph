package pipeline

import (
	"context"
	"os"
	"os/exec"
	"runtime"
)

// IsMac returns true if running on macOS
func IsMac(ctx context.Context) bool {
	return runtime.GOOS == "darwin"
}

// IsLinux returns true if running on Linux
func IsLinux(ctx context.Context) bool {
	return runtime.GOOS == "linux"
}

// CommandExists returns a condition that checks if a command is available
func CommandExists(cmd string) func(context.Context) bool {
	return func(ctx context.Context) bool {
		_, err := exec.LookPath(cmd)
		return err == nil
	}
}

// FileExists returns a condition that checks if a file or directory exists
func FileExists(path string) func(context.Context) bool {
	return func(ctx context.Context) bool {
		expandedPath := os.ExpandEnv(path)
		_, err := os.Stat(expandedPath)
		return err == nil
	}
}

// EnvSet returns a condition that checks if an environment variable is set
func EnvSet(key string) func(context.Context) bool {
	return func(ctx context.Context) bool {
		_, exists := os.LookupEnv(key)
		return exists
	}
}

// Not inverts a condition
func Not(condition func(context.Context) bool) func(context.Context) bool {
	return func(ctx context.Context) bool {
		return !condition(ctx)
	}
}

// And combines multiple conditions with AND logic
func And(conditions ...func(context.Context) bool) func(context.Context) bool {
	return func(ctx context.Context) bool {
		for _, cond := range conditions {
			if !cond(ctx) {
				return false
			}
		}
		return true
	}
}

// Or combines multiple conditions with OR logic
func Or(conditions ...func(context.Context) bool) func(context.Context) bool {
	return func(ctx context.Context) bool {
		for _, cond := range conditions {
			if cond(ctx) {
				return true
			}
		}
		return false
	}
}
