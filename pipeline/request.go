package pipeline

import (
	"os"
	"runtime"

	"github.com/cwood/dotgraph/exec"
	"github.com/cwood/dotgraph/pkg"
)

// Request holds all dependencies and configuration for stage execution.
// The type parameter T is the application-specific config type.
type Request[T any] struct {
	// Env contains runtime environment information
	Env Environment

	// Services contains injected service dependencies
	Services Services

	// Options contains execution options
	Options Options

	// Config is the application-specific configuration
	Config T
}

// Environment contains runtime environment information
type Environment struct {
	// OS is the operating system (e.g., "darwin", "linux")
	OS string

	// Arch is the architecture (e.g., "amd64", "arm64")
	Arch string

	// WorkDir is the base directory for file operations (typically $HOME)
	WorkDir string
}

// Services contains injected service dependencies
type Services struct {
	// Executor runs external commands
	Executor exec.CommandExecutor

	// Installer manages package installation
	Installer pkg.Manager
}

// Options contains execution options
type Options struct {
	// DryRun when true prevents actual changes
	DryRun bool

	// Verbose enables detailed logging
	Verbose bool
}

// NewEnvironment creates an Environment with default values from the runtime
func NewEnvironment() Environment {
	workDir := os.Getenv("HOME")
	if workDir == "" {
		workDir, _ = os.UserHomeDir()
	}
	return Environment{
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
		WorkDir: workDir,
	}
}

// NewServices creates Services with default real implementations
func NewServices(osName string) Services {
	return Services{
		Executor:  exec.NewRealExecutor(),
		Installer: pkg.NewManager(osName),
	}
}

// StageHandler is the function signature for stage handlers.
// The type parameter T matches the Request's config type.
type StageHandler[T any] func(*Request[T]) error
