package pipeline

import (
	"os"
	"path/filepath"
	"strings"
)

// Condition is a function that returns true if a condition is met
type Condition[T any] func(*Request[T]) bool

// FileExists returns a condition that checks if a file or directory exists.
// The path can contain $HOME which will be replaced with req.Env.WorkDir,
// or ~ which will be expanded to req.Env.WorkDir.
func FileExists[T any](path string) Condition[T] {
	return func(req *Request[T]) bool {
		expandedPath := expandPathWithWorkDir(path, req.Env.WorkDir)
		_, err := os.Stat(expandedPath)
		return err == nil
	}
}

// CommandExists returns a condition that checks if a command is available
// using the executor from the request's services.
func CommandExists[T any](cmd string) Condition[T] {
	return func(req *Request[T]) bool {
		_, err := req.Services.Executor.LookPath(cmd)
		return err == nil
	}
}

// EnvSet returns a condition that checks if an environment variable is set
func EnvSet[T any](key string) Condition[T] {
	return func(req *Request[T]) bool {
		_, exists := os.LookupEnv(key)
		return exists
	}
}

// IsMac returns a condition that checks if running on macOS
func IsMac[T any]() Condition[T] {
	return func(req *Request[T]) bool {
		return req.Env.OS == "darwin"
	}
}

// IsLinux returns a condition that checks if running on Linux
func IsLinux[T any]() Condition[T] {
	return func(req *Request[T]) bool {
		return req.Env.OS == "linux"
	}
}

// Not inverts a condition
func Not[T any](condition Condition[T]) Condition[T] {
	return func(req *Request[T]) bool {
		return !condition(req)
	}
}

// And combines multiple conditions with AND logic
func And[T any](conditions ...Condition[T]) Condition[T] {
	return func(req *Request[T]) bool {
		for _, cond := range conditions {
			if !cond(req) {
				return false
			}
		}
		return true
	}
}

// Or combines multiple conditions with OR logic
func Or[T any](conditions ...Condition[T]) Condition[T] {
	return func(req *Request[T]) bool {
		for _, cond := range conditions {
			if cond(req) {
				return true
			}
		}
		return false
	}
}

// expandPathWithWorkDir expands ~ and $HOME in a path using the provided workDir
func expandPathWithWorkDir(path, workDir string) string {
	// Replace $HOME with workDir
	path = strings.ReplaceAll(path, "$HOME", workDir)

	// Replace ~ at the start with workDir
	if strings.HasPrefix(path, "~") {
		path = workDir + path[1:]
	}

	// Expand other environment variables
	path = os.ExpandEnv(path)

	// Clean the path
	return filepath.Clean(path)
}
