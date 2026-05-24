package exec

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// verbose, when set, streams command output to the terminal live in addition
// to capturing it. Toggle with SetVerbose.
var verbose bool

// SetVerbose toggles live streaming of command output.
func SetVerbose(v bool) {
	verbose = v
}

// RunResult contains the result of running a command. Stdout and Stderr hold
// the captured output regardless of success.
type RunResult struct {
	Success bool
	Stdout  string
	Stderr  string
	Error   error
}

// Run executes a command and captures its output. When verbose is set the
// output also streams to the terminal live. On failure the captured output is
// printed to stderr so problems are never hidden in a log file.
func Run(name string, args ...string) RunResult {
	return run(name, args...)
}

func run(name string, args ...string) RunResult {
	cmd := exec.Command(name, args...)

	var stdout, stderr bytes.Buffer
	if verbose {
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	err := cmd.Run()
	result := RunResult{
		Success: err == nil,
		Stdout:  stdout.String(),
		Stderr:  stderr.String(),
		Error:   err,
	}

	if err != nil && !verbose {
		// In verbose mode the output already streamed live; otherwise surface
		// it now so a failure isn't silent.
		fmt.Fprintf(os.Stderr, "command failed: %s\n", strings.Join(append([]string{name}, args...), " "))
		io.WriteString(os.Stderr, result.Stdout)
		io.WriteString(os.Stderr, result.Stderr)
	}

	return result
}
