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

var logDir string

func init() {
	// Create log directory in user's home
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logDir = "/tmp/bootstrap-logs"
	} else {
		logDir = filepath.Join(homeDir, ".cache", "bootstrap-logs")
	}
	os.MkdirAll(logDir, 0755)
}

// RunResult contains the result of running a command
type RunResult struct {
	Success bool
	LogFile string
	Error   error
}

// Run executes a command and captures output
// On success: returns success with no log file
// On failure: writes output to log file and returns path
func Run(name string, arg ...string) RunResult {
	cmd := exec.Command(name, arg...)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		// Write failure log
		timestamp := time.Now().Format("20060102-150405")
		logFile := filepath.Join(logDir, fmt.Sprintf("%s-%s.log", name, timestamp))
		
		logContent := fmt.Sprintf("Command: %s %v\n", name, arg)
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

// RunQuiet executes a command silently, logging to file on error
func RunQuiet(name string, arg ...string) RunResult {
	return Run(name, arg...)
}

// RunWithOutput executes a command and shows output (deprecated - use RunQuiet instead)
func RunWithOutput(name string, arg ...string) error {
	result := Run(name, arg...)
	if !result.Success {
		return result.Error
	}
	return nil
}
