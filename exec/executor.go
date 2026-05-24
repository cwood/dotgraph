package exec

import "os/exec"

// CommandExecutor defines the interface for running commands.
type CommandExecutor interface {
	// Run executes a command and returns the result.
	Run(name string, args ...string) RunResult

	// LookPath searches for an executable in PATH.
	LookPath(cmd string) (string, error)
}

// RealExecutor implements CommandExecutor using os/exec.
type RealExecutor struct{}

// NewRealExecutor creates a new RealExecutor.
func NewRealExecutor() *RealExecutor {
	return &RealExecutor{}
}

// Run executes a command, capturing its output and—when verbose is set via
// SetVerbose—streaming it to the terminal live.
func (r *RealExecutor) Run(name string, args ...string) RunResult {
	return run(name, args...)
}

// LookPath searches for an executable in PATH.
func (r *RealExecutor) LookPath(cmd string) (string, error) {
	return exec.LookPath(cmd)
}
