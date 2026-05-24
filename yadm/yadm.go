package yadm

import (
	"github.com/cwood/dotgraph/exec"
)

// CloneOptions configures the clone behavior
type CloneOptions struct {
	Force bool
}

// Yadm provides operations for yadm dotfiles manager
type Yadm struct{}

// Clone clones a dotfiles repository using yadm
func (y *Yadm) Clone(repo string, opts *CloneOptions) exec.RunResult {
	args := []string{"clone", repo}

	if opts != nil && opts.Force {
		args = append(args, "-f")
	}

	return exec.Run("yadm", args...)
}

// Bootstrap runs the yadm bootstrap script
func (y *Yadm) Bootstrap() exec.RunResult {
	return exec.Run("yadm", "bootstrap")
}
