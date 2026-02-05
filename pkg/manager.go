package pkg

import (
	"fmt"
	"os/exec"
)

// Manager defines the interface for package managers
type Manager interface {
	Install(packages ...string) error
	IsInstalled(pkg string) bool
	Available() bool
	Name() string
}

// Package manager priority by OS
var managerPriority = map[string][]Manager{
	"darwin": {&Homebrew{}},
	"linux":  {&Yay{}, &Pacman{}},
}

// NewManager returns the first available package manager for the OS
func NewManager(os string) Manager {
	managers, ok := managerPriority[os]
	if !ok {
		return &Noop{}
	}

	for _, m := range managers {
		if m.Available() {
			return m
		}
	}

	return &Noop{}
}

// Noop is a no-op package manager for unsupported platforms
type Noop struct{}

func (n *Noop) Install(packages ...string) error {
	return fmt.Errorf("package manager not supported on this platform")
}

func (n *Noop) IsInstalled(pkg string) bool {
	return false
}

func (n *Noop) Available() bool {
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
