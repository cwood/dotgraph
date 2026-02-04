package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cwood/dotgraph/logger"
)

// Yay implements the Manager interface for Arch Linux yay
type Yay struct{}

// Install installs packages using yay (batch install)
// yay handles both pacman repos and AUR packages
func (y *Yay) Install(packages ...string) error {
	if len(packages) == 0 {
		return nil
	}

	if !commandExists("yay") {
		return fmt.Errorf("yay not installed")
	}

	logger.Info("Installing %d packages via yay: %s", len(packages), strings.Join(packages, ", "))

	args := append([]string{"-S", "--noconfirm"}, packages...)
	cmd := exec.Command("yay", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// IsInstalled checks if a package is installed via yay/pacman
func (y *Yay) IsInstalled(pkg string) bool {
	if !commandExists("yay") {
		return false
	}

	cmd := exec.Command("yay", "-Qi", pkg)
	return cmd.Run() == nil
}

// Name returns the name of the package manager
func (y *Yay) Name() string {
	return "yay"
}
