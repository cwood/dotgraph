package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

func init() {
	// Create a custom handler with clean formatting
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	
	handler := slog.NewTextHandler(os.Stdout, opts)
	Log = slog.New(handler)
}

// SetVerbose enables debug level logging
func SetVerbose(verbose bool) {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	if !verbose {
		opts.Level = slog.LevelInfo
	}
	
	handler := slog.NewTextHandler(os.Stdout, opts)
	Log = slog.New(handler)
}

// Info logs an info message
func Info(msg string, args ...any) {
	Log.Info(msg, args...)
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	Log.Debug(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	Log.Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	Log.Error(msg, args...)
}

// Success logs a success message with a checkmark
func Success(msg string, args ...any) {
	Log.Info("✓ "+msg, args...)
}

// Failure logs a failure message with an X
func Failure(msg string, args ...any) {
	Log.Error("✗ "+msg, args...)
}

// Stage logs a stage start message
func Stage(msg string, args ...any) {
	Log.Info("→ "+msg, args...)
}
