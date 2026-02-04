# dotgraph

[![Go Reference](https://pkg.go.dev/badge/github.com/cwood/dotgraph.svg)](https://pkg.go.dev/github.com/cwood/dotgraph)
[![Go Report Card](https://goreportcard.com/badge/github.com/cwood/dotgraph)](https://goreportcard.com/report/github.com/cwood/dotgraph)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go library for building graph-based system configuration and bootstrap scripts with dependency management, platform-specific execution, and clean logging.

## Features

- **Graph-based execution**: Define stages with dependencies that execute in parallel when possible
- **Platform-specific stages**: Automatically filter stages by OS (darwin/linux)
- **Conditional execution**: Skip stages based on conditions (command exists, file exists, etc.)
- **Package manager abstraction**: Unified interface for Homebrew (macOS) and yay (Arch Linux)
- **Clean logging**: Structured logging with slog, minimal output, failure logs saved to files
- **Command execution**: Execute shell commands with automatic error logging

## Installation

```bash
go get github.com/cwood/dotgraph
```

## Quick Start

```go
package main

import (
    "context"
    "github.com/cwood/dotgraph/pipeline"
    "github.com/cwood/dotgraph/logger"
)

func main() {
    logger.SetVerbose(false)
    
    graph := pipeline.NewGraph()
    
    // Add a basic stage
    git := graph.AddStage("update-git-submodules", func(ctx context.Context) error {
        // Your code here
        return nil
    })
    
    // Add platform-specific stages
    mac := graph.AddPlatform("darwin")
    mac.AddStage("install-homebrew-packages", func(ctx context.Context) error {
        // macOS-specific code
        return nil
    }).After(git).Unless(pipeline.CommandExists("brew"))
    
    linux := graph.AddPlatform("linux")
    linux.AddStage("install-yay-packages", func(ctx context.Context) error {
        // Linux-specific code
        return nil
    }).After(git).Unless(pipeline.CommandExists("yay"))
    
    // Execute the graph
    ctx := context.Background()
    if err := graph.Execute(ctx); err != nil {
        logger.Error("Bootstrap failed", "error", err)
        os.Exit(1)
    }
}
```

## Core Concepts

### Graph

The `Graph` manages stages and their dependencies, executing them in parallel when possible while respecting dependency order.

```go
graph := pipeline.NewGraph()
```

### Stages

Stages are individual tasks that can have dependencies and conditions:

```go
stage := graph.AddStage("stage-name", func(ctx context.Context) error {
    // Your code here
    return nil
})

// Add dependencies
stage.After(otherStage1, otherStage2)

// Skip if condition is true
stage.Unless(pipeline.CommandExists("git"))

// Require a command to exist
stage.Requires("git")

// Mark as optional (won't fail the graph)
stage.Optional()
```

### Platform-Specific Stages

Create stages that only run on specific platforms:

```go
mac := graph.AddPlatform("darwin")
mac.AddStage("macos-only", func(ctx context.Context) error {
    return nil
})

linux := graph.AddPlatform("linux")
linux.AddStage("linux-only", func(ctx context.Context) error {
    return nil
})
```

### Merge Points

Wait for multiple stages before continuing:

```go
merge := graph.AddMerge("wait-for-both", stage1, stage2)
merge.AddStage("after-both", func(ctx context.Context) error {
    return nil
})
```

### Conditions

Built-in conditions for common checks:

```go
pipeline.IsMac(ctx)                    // Running on macOS
pipeline.IsLinux(ctx)                  // Running on Linux
pipeline.CommandExists("git")          // Command in PATH
pipeline.FileExists("~/.config/file")  // File exists
pipeline.EnvSet("HOME")                // Environment variable set
pipeline.Not(condition)                // Invert condition
pipeline.And(cond1, cond2)            // All conditions true
pipeline.Or(cond1, cond2)             // Any condition true
```

### Package Managers

Unified interface for system package managers:

```go
import "github.com/cwood/dotgraph/pkg"

mgr := pkg.NewManager(runtime.GOOS)

// Install packages
mgr.Install("git", "vim", "tmux")

// Check if installed
if mgr.IsInstalled("git") {
    // ...
}

// Homebrew-specific: run brew bundle
if homebrew, ok := mgr.(*pkg.Homebrew); ok {
    homebrew.Bundle("~/Brewfile")
}
```

### Command Execution

Execute commands with automatic error logging:

```go
import "github.com/cwood/dotgraph/exec"

result := exec.RunQuiet("git", "clone", "https://...")
if !result.Success {
    logger.Failure("Git clone failed", "log", result.LogFile)
    return result.Error
}
```

### Logging

Structured logging with clean output:

```go
import "github.com/cwood/dotgraph/logger"

logger.SetVerbose(true)  // Enable debug logs

logger.Info("Starting bootstrap")
logger.Debug("Checking dependencies")
logger.Success("Completed successfully")
logger.Failure("Failed to install", "error", err)
logger.Stage("Installing packages")
```

## Example: Bootstrap Script

See the [examples](examples/) directory for complete bootstrap script examples.

## License

MIT License - see [LICENSE](LICENSE) for details.
