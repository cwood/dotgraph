// Package logger provides simple, human-readable output for the bootstrap CLI.
//
// Every function takes a printf-style format string and arguments. Messages
// print as clean lines with no timestamp or level prefix; Success, Failure,
// and Stage prepend a glyph. Debug output is suppressed unless SetVerbose(true)
// has been called.
package logger

import (
	"fmt"
	"os"
)

var verbose bool

// SetVerbose toggles whether Debug messages are printed.
func SetVerbose(v bool) {
	verbose = v
}

// emit writes a single formatted line to stdout. With no arguments the format
// is printed verbatim, so a stray '%' in a bare message can't be misread as a
// verb.
func emit(format string, args ...any) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stdout, format)
		return
	}
	fmt.Fprintf(os.Stdout, format+"\n", args...)
}

// Info prints an informational message.
func Info(format string, args ...any) {
	emit(format, args...)
}

// Debug prints a message only when verbose output is enabled.
func Debug(format string, args ...any) {
	if verbose {
		emit(format, args...)
	}
}

// Warn prints a warning message.
func Warn(format string, args ...any) {
	emit(format, args...)
}

// Error prints an error message.
func Error(format string, args ...any) {
	emit(format, args...)
}

// Success prints a message prefixed with a check mark.
func Success(format string, args ...any) {
	emit("✓ "+format, args...)
}

// Failure prints a message prefixed with a cross mark.
func Failure(format string, args ...any) {
	emit("✗ "+format, args...)
}

// Stage prints a stage header prefixed with an arrow.
func Stage(format string, args ...any) {
	emit("→ "+format, args...)
}
