package exec

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// CommandExecutor defines the interface for running commands
type CommandExecutor interface {
	// Run executes a command and returns the result
	Run(name string, args ...string) RunResult

	// LookPath searches for an executable in PATH
	LookPath(cmd string) (string, error)
}

// RealExecutor implements CommandExecutor using os/exec
type RealExecutor struct {
	// LogDir is the directory where failure logs are written
	// If empty, defaults to ~/.cache/bootstrap-logs or /tmp/bootstrap-logs
	LogDir string
}

// NewRealExecutor creates a new RealExecutor with default log directory
func NewRealExecutor() *RealExecutor {
	logDir := "/tmp/bootstrap-logs"
	if homeDir, err := os.UserHomeDir(); err == nil {
		logDir = filepath.Join(homeDir, ".cache", "bootstrap-logs")
	}
	os.MkdirAll(logDir, 0755)
	return &RealExecutor{LogDir: logDir}
}

// Run executes a command and captures output
// On success: returns success with no log file
// On failure: writes output to log file and returns path
func (r *RealExecutor) Run(name string, args ...string) RunResult {
	cmd := exec.Command(name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		// Write failure log
		timestamp := time.Now().Format("20060102-150405")
		logFile := filepath.Join(r.LogDir, fmt.Sprintf("%s-%s.log", name, timestamp))

		logContent := fmt.Sprintf("Command: %s %v\n", name, args)
		logContent += fmt.Sprintf("Exit Code: %v\n", err)
		logContent += fmt.Sprintf("\n=== STDOUT ===\n%s\n", stdout.String())
		logContent += fmt.Sprintf("\n=== STDERR ===\n%s\n", stderr.String())

		if writeErr := os.WriteFile(logFile, []byte(logContent), 0644); writeErr != nil {
			log.Printf("Failed to write log file: %v", writeErr)
			return RunResult{Success: false, Error: err}
		}

		return RunResult{Success: false, LogFile: logFile, Error: err}
	}

	return RunResult{Success: true}
}

// LookPath searches for an executable in PATH
func (r *RealExecutor) LookPath(cmd string) (string, error) {
	return exec.LookPath(cmd)
}
