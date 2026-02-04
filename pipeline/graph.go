package pipeline

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/cwood/dotgraph/logger"
)

// Graph represents a dependency graph of stages
type Graph struct {
	stages   map[string]*GraphStage
	platform string
}

// GraphStage represents a stage in the dependency graph
type GraphStage struct {
	name         string
	run          func(ctx context.Context) error
	dependencies []*GraphStage
	platform     string // Empty means all platforms
	requires     []string
	unless       []func(context.Context) bool
	optional     bool
	executed     bool
	mu           sync.Mutex
}

// NewGraph creates a new dependency graph
func NewGraph() *Graph {
	return &Graph{
		stages:   make(map[string]*GraphStage),
		platform: runtime.GOOS,
	}
}

// AddStage adds a stage to the graph
func (g *Graph) AddStage(name string, run func(context.Context) error) *GraphStage {
	stage := &GraphStage{
		name:         name,
		run:          run,
		dependencies: make([]*GraphStage, 0),
		requires:     make([]string, 0),
		unless:       make([]func(context.Context) bool, 0),
	}
	g.stages[name] = stage
	return stage
}

// AddPlatform creates a platform-specific stage builder
func (g *Graph) AddPlatform(platform string) *PlatformBuilder {
	return &PlatformBuilder{
		graph:    g,
		platform: platform,
	}
}

// AddMerge creates a merge node that waits for multiple stages
func (g *Graph) AddMerge(name string, stages ...*GraphStage) *MergeBuilder {
	// Create a no-op stage that just waits for dependencies
	merge := &GraphStage{
		name:         name,
		run:          func(ctx context.Context) error { return nil },
		dependencies: stages,
		requires:     make([]string, 0),
		unless:       make([]func(context.Context) bool, 0),
	}
	g.stages[name] = merge
	return &MergeBuilder{
		graph:      g,
		mergeStage: merge,
	}
}

// MergeBuilder helps build stages after a merge point
type MergeBuilder struct {
	graph      *Graph
	mergeStage *GraphStage
}

// AddStage adds a stage that runs after the merge
func (mb *MergeBuilder) AddStage(name string, run func(context.Context) error) *GraphStage {
	stage := mb.graph.AddStage(name, run)
	stage.dependencies = append(stage.dependencies, mb.mergeStage)
	return stage
}

// Execute runs the graph, respecting dependencies
func (g *Graph) Execute(ctx context.Context) error {
	logger.Info("Executing bootstrap graph", "stages", len(g.stages))

	// Find root stages (no dependencies)
	roots := make([]*GraphStage, 0)
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
		go func(s *GraphStage) {
			defer wg.Done()
			if err := g.executeStage(ctx, s); err != nil {
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
func (g *Graph) executeStage(ctx context.Context, stage *GraphStage) error {
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
		if condition(ctx) {
			logger.Debug("Skipping stage", "stage", stage.name, "reason", "unless condition met")
			stage.mu.Lock()
			stage.executed = true
			stage.mu.Unlock()
			return nil
		}
	}

	// Check required commands
	for _, cmd := range stage.requires {
		if !CommandExists(cmd)(ctx) {
			logger.Debug("Skipping stage", "stage", stage.name, "reason", "missing requirement", "command", cmd)
			stage.mu.Lock()
			stage.executed = true
			stage.mu.Unlock()
			return nil
		}
	}

	// Execute the stage
	logger.Stage(stage.name)
	if err := stage.run(ctx); err != nil {
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
			go func(s *GraphStage) {
				defer wg.Done()
				if err := g.executeStage(ctx, s); err != nil {
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
func (g *Graph) findDependents(stage *GraphStage) []*GraphStage {
	dependents := make([]*GraphStage, 0)
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
func (g *Graph) allDependenciesMet(stage *GraphStage) bool {
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
func (s *GraphStage) After(stages ...*GraphStage) *GraphStage {
	s.dependencies = append(s.dependencies, stages...)
	return s
}

// Requires adds a command requirement (must exist in PATH)
func (s *GraphStage) Requires(cmd string) *GraphStage {
	s.requires = append(s.requires, cmd)
	return s
}

// Unless adds a condition that skips the stage if true
func (s *GraphStage) Unless(condition func(context.Context) bool) *GraphStage {
	s.unless = append(s.unless, condition)
	return s
}

// Optional marks the stage as optional (won't fail the graph)
func (s *GraphStage) Optional() *GraphStage {
	s.optional = true
	return s
}

// PlatformBuilder helps build platform-specific stages
type PlatformBuilder struct {
	graph    *Graph
	platform string
}

// AddStage adds a platform-specific stage
func (pb *PlatformBuilder) AddStage(name string, run func(context.Context) error) *GraphStage {
	stage := pb.graph.AddStage(name, run)
	stage.platform = pb.platform
	return stage
}
