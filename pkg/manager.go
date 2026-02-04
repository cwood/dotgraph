package pkg

import (
	"fmt"
	"os/exec"
)

// Manager defines the interface for package managers
type Manager interface {
	Install(packages ...string) error
	IsInstalled(pkg string) bool
	Name() string
}

// NewManager creates a package manager based on the OS
func NewManager(os string) Manager {
	switch os {
	case "darwin":
		return &Homebrew{}
	case "linux":
		return &Yay{}
	default:
		return &Noop{}
	}
}

// Noop is a no-op package manager for unsupported platforms
type Noop struct{}

func (n *Noop) Install(packages ...string) error {
	return fmt.Errorf("package manager not supported on this platform")
}

func (n *Noop) IsInstalled(pkg string) bool {
	return false
}

func (n *Noop) Name() string {
	return "noop"
}

// commandExists checks if a command is available in PATH
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
