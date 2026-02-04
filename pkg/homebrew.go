package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	dgexec "github.com/cwood/dotgraph/exec"
	"github.com/cwood/dotgraph/logger"
)

// Homebrew implements the Manager interface for macOS Homebrew
type Homebrew struct{}

// Install installs packages using Homebrew (batch install)
func (h *Homebrew) Install(packages ...string) error {
	if len(packages) == 0 {
		return nil
	}

	if !commandExists("brew") {
		return fmt.Errorf("homebrew not installed")
	}

	logger.Info("Installing %d packages via Homebrew: %s", len(packages), strings.Join(packages, ", "))

	args := append([]string{"install"}, packages...)
	cmd := exec.Command("brew", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// IsInstalled checks if a package is installed via Homebrew
func (h *Homebrew) IsInstalled(pkg string) bool {
	if !commandExists("brew") {
		return false
	}

	cmd := exec.Command("brew", "list", pkg)
	return cmd.Run() == nil
}

// Name returns the name of the package manager
func (h *Homebrew) Name() string {
	return "homebrew"
}

// Bundle runs brew bundle with the specified Brewfile
func (h *Homebrew) Bundle(brewfilePath string) error {
	if !commandExists("brew") {
		return fmt.Errorf("homebrew not installed")
	}

	expandedPath := os.ExpandEnv(brewfilePath)
	
	result := dgexec.RunQuiet("brew", "bundle", "--file="+expandedPath)
	if result.Success {
		logger.Info("  ✓ Brewfile packages installed")
		return nil
	}
	logger.Info("  ✗ Failed to install Brewfile packages - see log: %s", result.LogFile)
	return result.Error
}
