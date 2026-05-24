package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cwood/dotgraph/logger"
)

// Pacman implements the Manager interface for Arch Linux pacman
type Pacman struct{}

// Install installs packages using pacman
func (p *Pacman) Install(packages ...string) error {
	if len(packages) == 0 {
		return nil
	}

	if !commandExists("pacman") {
		return fmt.Errorf("pacman not installed")
	}

	logger.Info("Installing %d packages via pacman: %s", len(packages), strings.Join(packages, ", "))

	args := append([]string{"-S", "--noconfirm"}, packages...)
	cmd := exec.Command("sudo", append([]string{"pacman"}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// IsInstalled checks if a package is installed via pacman
func (p *Pacman) IsInstalled(pkg string) bool {
	if !commandExists("pacman") {
		return false
	}

	cmd := exec.Command("pacman", "-Qi", pkg)
	return cmd.Run() == nil
}

// Available checks if pacman is installed
func (p *Pacman) Available() bool {
	return commandExists("pacman")
}

// Name returns the name of the package manager
func (p *Pacman) Name() string {
	return "pacman"
}
