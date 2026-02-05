package pipeline

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/cwood/dotgraph/logger"
)

// Graph represents a dependency graph of stages.
// The type parameter T is the application-specific config type.
type Graph[T any] struct {
	stages   map[string]*GraphStage[T]
	platform string
}

// GraphStage represents a stage in the dependency graph.
// The type parameter T matches the Graph's config type.
type GraphStage[T any] struct {
	name         string
	run          StageHandler[T]
	dependencies []*GraphStage[T]
	platform     string // Empty means all platforms
	requires     []string
	unless       []func(*Request[T]) bool
	optional     bool
	executed     bool
	mu           sync.Mutex
}

// NewGraph creates a new dependency graph
func NewGraph[T any]() *Graph[T] {
	return &Graph[T]{
		stages:   make(map[string]*GraphStage[T]),
		platform: runtime.GOOS,
	}
}

// AddStage adds a stage to the graph
func (g *Graph[T]) AddStage(name string, run StageHandler[T]) *GraphStage[T] {
	stage := &GraphStage[T]{
		name:         name,
		run:          run,
		dependencies: make([]*GraphStage[T], 0),
		requires:     make([]string, 0),
		unless:       make([]func(*Request[T]) bool, 0),
	}
	g.stages[name] = stage
	return stage
}

// AddPlatform creates a platform-specific stage builder
func (g *Graph[T]) AddPlatform(platform string) *PlatformBuilder[T] {
	return &PlatformBuilder[T]{
		graph:    g,
		platform: platform,
	}
}

// AddMerge creates a merge node that waits for multiple stages
func (g *Graph[T]) AddMerge(name string, stages ...*GraphStage[T]) *MergeBuilder[T] {
	// Create a no-op stage that just waits for dependencies
	merge := &GraphStage[T]{
		name:         name,
		run:          func(req *Request[T]) error { return nil },
		dependencies: stages,
		requires:     make([]string, 0),
		unless:       make([]func(*Request[T]) bool, 0),
	}
	g.stages[name] = merge
	return &MergeBuilder[T]{
		graph:      g,
		mergeStage: merge,
	}
}

// MergeBuilder helps build stages after a merge point
type MergeBuilder[T any] struct {
	graph      *Graph[T]
	mergeStage *GraphStage[T]
}

// AddStage adds a stage that runs after the merge
func (mb *MergeBuilder[T]) AddStage(name string, run StageHandler[T]) *GraphStage[T] {
	stage := mb.graph.AddStage(name, run)
	stage.dependencies = append(stage.dependencies, mb.mergeStage)
	return stage
}

// Execute runs the graph, respecting dependencies
func (g *Graph[T]) Execute(ctx context.Context, req *Request[T]) error {
	logger.Info("Executing bootstrap graph", "stages", len(g.stages))

	// Find root stages (no dependencies)
	roots := make([]*GraphStage[T], 0)
	for _, stage := range g.stages {
		if len(stage.dependencies) == 0 {
			roots = append(roots, stage)
		}
	}

	// Execute from roots
	var wg sync.WaitGroup
	errChan := make(chan error, len(g.stages))

	for _, root := range roots {
		wg.Add(1)
		go func(s *GraphStage[T]) {
			defer wg.Done()
			if err := g.executeStage(ctx, req, s); err != nil {
				errChan <- err
			}
		}(root)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	logger.Success("Bootstrap graph completed successfully")
	return nil
}

// executeStage executes a stage and its dependents
func (g *Graph[T]) executeStage(ctx context.Context, req *Request[T], stage *GraphStage[T]) error {
	stage.mu.Lock()
	if stage.executed {
		stage.mu.Unlock()
		return nil
	}
	stage.mu.Unlock()

	// Check platform
	if stage.platform != "" && stage.platform != g.platform {
		logger.Debug("Skipping stage", "stage", stage.name, "reason", "platform mismatch", "expected", stage.platform, "current", g.platform)
		stage.mu.Lock()
		stage.executed = true
		stage.mu.Unlock()
		return nil
	}

	// Check unless conditions
	for _, condition := range stage.unless {
		if condition(req) {
			logger.Debug("Skipping stage", "stage", stage.name, "reason", "unless condition met")
			stage.mu.Lock()
			stage.executed = true
			stage.mu.Unlock()
			return nil
		}
	}

	// Check required commands
	for _, cmd := range stage.requires {
		_, err := req.Services.Executor.LookPath(cmd)
		if err != nil {
			if stage.optional {
				logger.Debug("Skipping stage", "stage", stage.name, "reason", "missing requirement", "command", cmd)
				stage.mu.Lock()
				stage.executed = true
				stage.mu.Unlock()
				return nil
			}
			return fmt.Errorf("stage %s requires command %s which is not available", stage.name, cmd)
		}
	}

	// Execute the stage
	logger.Stage(stage.name)
	if err := stage.run(req); err != nil {
		if stage.optional {
			logger.Warn("Stage failed (optional)", "stage", stage.name, "error", err)
		} else {
			return fmt.Errorf("stage %s failed: %w", stage.name, err)
		}
	} else {
		logger.Success(stage.name)
	}

	stage.mu.Lock()
	stage.executed = true
	stage.mu.Unlock()

	// Find and execute dependent stages
	dependents := g.findDependents(stage)
	if len(dependents) > 0 {
		var wg sync.WaitGroup
		errChan := make(chan error, len(dependents))

		for _, dep := range dependents {
			// Check if all dependencies are satisfied
			if !g.allDependenciesMet(dep) {
				continue
			}

			wg.Add(1)
			go func(s *GraphStage[T]) {
				defer wg.Done()
				if err := g.executeStage(ctx, req, s); err != nil {
					errChan <- err
				}
			}(dep)
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// findDependents finds stages that depend on the given stage
func (g *Graph[T]) findDependents(stage *GraphStage[T]) []*GraphStage[T] {
	dependents := make([]*GraphStage[T], 0)
	for _, s := range g.stages {
		for _, dep := range s.dependencies {
			if dep == stage {
				dependents = append(dependents, s)
				break
			}
		}
	}
	return dependents
}

// allDependenciesMet checks if all dependencies of a stage are executed
func (g *Graph[T]) allDependenciesMet(stage *GraphStage[T]) bool {
	for _, dep := range stage.dependencies {
		dep.mu.Lock()
		executed := dep.executed
		dep.mu.Unlock()
		if !executed {
			return false
		}
	}
	return true
}

// After adds dependencies to this stage
func (s *GraphStage[T]) After(stages ...*GraphStage[T]) *GraphStage[T] {
	s.dependencies = append(s.dependencies, stages...)
	return s
}

// Requires adds a command requirement (must exist in PATH)
func (s *GraphStage[T]) Requires(cmd string) *GraphStage[T] {
	s.requires = append(s.requires, cmd)
	return s
}

// Unless adds a condition that skips the stage if true
func (s *GraphStage[T]) Unless(condition func(*Request[T]) bool) *GraphStage[T] {
	s.unless = append(s.unless, condition)
	return s
}

// Optional marks the stage as optional (won't fail the graph)
func (s *GraphStage[T]) Optional() *GraphStage[T] {
	s.optional = true
	return s
}

// PlatformBuilder helps build platform-specific stages
type PlatformBuilder[T any] struct {
	graph    *Graph[T]
	platform string
}

// AddStage adds a platform-specific stage
func (pb *PlatformBuilder[T]) AddStage(name string, run StageHandler[T]) *GraphStage[T] {
	stage := pb.graph.AddStage(name, run)
	stage.platform = pb.platform
	return stage
}
