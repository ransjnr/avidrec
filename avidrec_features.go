package main

import (
	"fmt"
	"math"
)

// DependencyFeatures represents ML features for a single dependency
type DependencyFeatures struct {
	// Dependency identifiers
	SourceModule string
	TargetModule string

	// Code complexity features
	SourceLOC        int     // Lines of code in source module
	TargetLOC        int     // Lines of code in target module
	SourceCyclomatic float64 // Cyclomatic complexity of source
	TargetCyclomatic float64 // Cyclomatic complexity of target

	// Change frequency features (0-100 scale)
	SourceCommits      int // Commits in last 30 days
	TargetCommits      int // Commits in last 30 days
	SourceChangeRate   float64 // Files changed / total files (%)
	TargetChangeRate   float64 // Files changed / total files (%)

	// Dependency structure features
	SourceInDegree  int // How many modules depend on source
	SourceOutDegree int // How many modules source depends on
	TargetInDegree  int // How many modules depend on target
	TargetOutDegree int // How many modules target depends on

	// Team/ownership features
	SourceTeams    int // Number of teams owning source module
	TargetTeams    int // Number of teams owning target module
	SharedTeams    int // Overlapping teams

	// Architectural layer features (0-5 scale)
	SourceLayer int // Layer distance from foundation (0=foundation, 5=top)
	TargetLayer int // Layer distance from foundation

	// Historical violation feature
	IsViolation int // 1 if this dependency is a violation, 0 otherwise (LABEL)
}

// FeatureExtractor extracts ML features from a dependency graph
type FeatureExtractor struct {
	graph *DependencyGraph
	// Cache for computed features
	violations map[string]map[string]bool // violations[source][target] = true if violation
}

// NewFeatureExtractor creates a new feature extractor
func NewFeatureExtractor(graph *DependencyGraph) *FeatureExtractor {
	return &FeatureExtractor{
		graph:      graph,
		violations: make(map[string]map[string]bool),
	}
}

// ExtractFeaturesForDependency extracts features for a single dependency
func (fe *FeatureExtractor) ExtractFeaturesForDependency(source, target string) *DependencyFeatures {
	features := &DependencyFeatures{
		SourceModule: source,
		TargetModule: target,
	}

	// Extract code complexity features
	sourceModule := fe.graph.Modules[source]
	targetModule := fe.graph.Modules[target]

	if sourceModule != nil {
		features.SourceLOC = sourceModule.LOC
	}
	if targetModule != nil {
		features.TargetLOC = targetModule.LOC
	}

	// Estimate cyclomatic complexity from LOC (simple heuristic)
	features.SourceCyclomatic = fe.estimateCyclomatic(features.SourceLOC)
	features.TargetCyclomatic = fe.estimateCyclomatic(features.TargetLOC)

	// Extract degree features
	features.SourceInDegree = fe.graph.InDegree[source]
	features.SourceOutDegree = fe.graph.OutDegree[source]
	features.TargetInDegree = fe.graph.InDegree[target]
	features.TargetOutDegree = fe.graph.OutDegree[target]

	// Extract layer features (based on topological level)
	features.SourceLayer = fe.estimateLayer(source)
	features.TargetLayer = fe.estimateLayer(target)

	// Set label (0 or 1)
	features.IsViolation = 0 // Will be set during dataset generation

	return features
}

// estimateCyclomatic estimates cyclomatic complexity from LOC
// Simple heuristic: complexity ~ sqrt(LOC) / 10
func (fe *FeatureExtractor) estimateCyclomatic(loc int) float64 {
	if loc == 0 {
		return 1.0
	}
	return math.Sqrt(float64(loc)) / 2.0
}

// estimateLayer estimates the architectural layer of a module
// Layer 0 (foundation): modules with no outgoing dependencies
// Higher layers depend on lower layers
func (fe *FeatureExtractor) estimateLayer(moduleID string) int {
	return fe.calculateLayerBFS(moduleID)
}

// calculateLayerBFS calculates layer using breadth-first search
func (fe *FeatureExtractor) calculateLayerBFS(moduleID string) int {
	visited := make(map[string]int) // moduleID -> layer
	queue := []string{moduleID}
	visited[moduleID] = 0

	maxLayer := 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Get all modules that current depends on
		deps := fe.graph.GetDependencies(current)
		for _, dep := range deps {
			if _, seen := visited[dep]; !seen {
				visited[dep] = visited[current] + 1
				queue = append(queue, dep)
				if visited[dep] > maxLayer {
					maxLayer = visited[dep]
				}
			}
		}
	}

	return maxLayer
}

// ExtractAllFeatures extracts features for all dependencies
func (fe *FeatureExtractor) ExtractAllFeatures() []*DependencyFeatures {
	features := make([]*DependencyFeatures, 0)

	for _, edge := range fe.graph.Edges {
		feat := fe.ExtractFeaturesForDependency(edge.Source, edge.Target)
		features = append(features, feat)
	}

	return features
}

// NormalizeFeatures normalizes features to 0-1 range
func (fe *FeatureExtractor) NormalizeFeatures(features []*DependencyFeatures) {
	if len(features) == 0 {
		return
	}

	// Find min/max for each feature
	maxSourceLOC := 0
	maxTargetLOC := 0
	maxSourceCommits := 0
	maxTargetCommits := 0
	maxSourceInDegree := 0
	maxSourceOutDegree := 0
	maxTargetInDegree := 0
	maxTargetOutDegree := 0

	for _, f := range features {
		if f.SourceLOC > maxSourceLOC {
			maxSourceLOC = f.SourceLOC
		}
		if f.TargetLOC > maxTargetLOC {
			maxTargetLOC = f.TargetLOC
		}
		if f.SourceCommits > maxSourceCommits {
			maxSourceCommits = f.SourceCommits
		}
		if f.TargetCommits > maxTargetCommits {
			maxTargetCommits = f.TargetCommits
		}
		if f.SourceInDegree > maxSourceInDegree {
			maxSourceInDegree = f.SourceInDegree
		}
		if f.SourceOutDegree > maxSourceOutDegree {
			maxSourceOutDegree = f.SourceOutDegree
		}
		if f.TargetInDegree > maxTargetInDegree {
			maxTargetInDegree = f.TargetInDegree
		}
		if f.TargetOutDegree > maxTargetOutDegree {
			maxTargetOutDegree = f.TargetOutDegree
		}
	}

	// Normalize
	for _, f := range features {
		if maxSourceLOC > 0 {
			_ = float64(f.SourceLOC) / float64(maxSourceLOC)
		}
		if maxTargetLOC > 0 {
			_ = float64(f.TargetLOC) / float64(maxTargetLOC)
		}
		if maxSourceCommits > 0 {
			f.SourceChangeRate = float64(f.SourceCommits) / float64(maxSourceCommits)
		}
		if maxTargetCommits > 0 {
			f.TargetChangeRate = float64(f.TargetCommits) / float64(maxTargetCommits)
		}
	}
}

// PrintFeatures prints features in readable format
func (fe *FeatureExtractor) PrintFeatures(features []*DependencyFeatures) {
	fmt.Println("\n=== Extracted Features ===")
	for i, f := range features {
		if i > 5 { // Only print first 5
			fmt.Printf("... and %d more\n", len(features)-5)
			break
		}
		fmt.Printf("\nDependency %d: %s → %s\n", i+1, f.SourceModule, f.TargetModule)
		fmt.Printf("  Source LOC: %d, Cyclomatic: %.2f\n", f.SourceLOC, f.SourceCyclomatic)
		fmt.Printf("  Target LOC: %d, Cyclomatic: %.2f\n", f.TargetLOC, f.TargetCyclomatic)
		fmt.Printf("  Source Degree: in=%d, out=%d | Target Degree: in=%d, out=%d\n",
			f.SourceInDegree, f.SourceOutDegree, f.TargetInDegree, f.TargetOutDegree)
		fmt.Printf("  Layers: source=%d, target=%d\n", f.SourceLayer, f.TargetLayer)
		fmt.Printf("  Label (Violation): %d\n", f.IsViolation)
	}
	fmt.Println()
}

// GetFeatureNames returns the names of all features in order
func GetFeatureNames() []string {
	return []string{
		"SourceLOC",
		"TargetLOC",
		"SourceCyclomatic",
		"TargetCyclomatic",
		"SourceCommits",
		"TargetCommits",
		"SourceChangeRate",
		"TargetChangeRate",
		"SourceInDegree",
		"SourceOutDegree",
		"TargetInDegree",
		"TargetOutDegree",
		"SourceTeams",
		"TargetTeams",
		"SharedTeams",
		"SourceLayer",
		"TargetLayer",
	}
}

// FeaturesToArray converts feature struct to float64 array for ML
func FeaturesToArray(f *DependencyFeatures) []float64 {
	return []float64{
		float64(f.SourceLOC),
		float64(f.TargetLOC),
		f.SourceCyclomatic,
		f.TargetCyclomatic,
		float64(f.SourceCommits),
		float64(f.TargetCommits),
		f.SourceChangeRate,
		f.TargetChangeRate,
		float64(f.SourceInDegree),
		float64(f.SourceOutDegree),
		float64(f.TargetInDegree),
		float64(f.TargetOutDegree),
		float64(f.SourceTeams),
		float64(f.TargetTeams),
		float64(f.SharedTeams),
		float64(f.SourceLayer),
		float64(f.TargetLayer),
	}
}