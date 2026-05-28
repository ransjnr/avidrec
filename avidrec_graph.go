package main

import (
	"fmt"
)

// Module represents a code module/package in the dependency graph
type Module struct {
	ID       string // Unique identifier (e.g., "github.com/user/project/pkg")
	Name     string // Human-readable name (e.g., "auth", "db", "web")
	FilePath string // Path to the module in codebase
	LOC      int    // Lines of code (for analysis)
}

// Dependency represents a single dependency edge in the graph
type Dependency struct {
	Source     string // Module that depends (A)
	Target     string // Module being depended on (B) - A depends on B
	Weight     int    // Strength of dependency (usage count)
	ImportLine int    // Line number where import occurs
}

// DependencyGraph represents the entire module dependency graph
type DependencyGraph struct {
	Modules      map[string]*Module           // Module ID -> Module
	Edges        []Dependency                 // All edges/dependencies
	AdjList      map[string][]string          // Adjacency list: Module -> List of modules it depends on
	ReverseAdj   map[string][]string          // Reverse adjacency list: Module -> Modules that depend on it
	InDegree     map[string]int               // In-degree of each node
	OutDegree    map[string]int               // Out-degree of each node
}

// NewDependencyGraph creates and initializes a new empty dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		Modules:    make(map[string]*Module),
		Edges:      make([]Dependency, 0),
		AdjList:    make(map[string][]string),
		ReverseAdj: make(map[string][]string),
		InDegree:   make(map[string]int),
		OutDegree:  make(map[string]int),
	}
}

// AddModule adds a module to the graph
func (g *DependencyGraph) AddModule(module *Module) {
	if _, exists := g.Modules[module.ID]; !exists {
		g.Modules[module.ID] = module
		g.AdjList[module.ID] = make([]string, 0)
		g.ReverseAdj[module.ID] = make([]string, 0)
		g.InDegree[module.ID] = 0
		g.OutDegree[module.ID] = 0
	}
}

// AddDependency adds a dependency edge from source to target
// If the dependency already exists, increase its weight
func (g *DependencyGraph) AddDependency(source, target string, weight int) error {
	// Ensure both modules exist
	if _, exists := g.Modules[source]; !exists {
		return fmt.Errorf("source module '%s' not found in graph", source)
	}
	if _, exists := g.Modules[target]; !exists {
		return fmt.Errorf("target module '%s' not found in graph", target)
	}

	// Prevent self-loops for clarity (though they can exist)
	if source == target {
		return fmt.Errorf("self-loops not allowed: '%s' cannot depend on itself", source)
	}

	// Check if edge already exists
	for i, edge := range g.Edges {
		if edge.Source == source && edge.Target == target {
			// Edge exists, increase weight
			g.Edges[i].Weight += weight
			return nil
		}
	}

	// Add new edge
	dep := Dependency{
		Source: source,
		Target: target,
		Weight: weight,
	}
	g.Edges = append(g.Edges, dep)

	// Update adjacency lists
	g.AdjList[source] = append(g.AdjList[source], target)
	g.ReverseAdj[target] = append(g.ReverseAdj[target], source)

	// Update degrees
	g.OutDegree[source]++
	g.InDegree[target]++

	return nil
}

// GetDependencies returns all modules that 'module' depends on
func (g *DependencyGraph) GetDependencies(moduleID string) []string {
	if deps, exists := g.AdjList[moduleID]; exists {
		return deps
	}
	return []string{}
}

// GetDependents returns all modules that depend on 'module'
func (g *DependencyGraph) GetDependents(moduleID string) []string {
	if dependents, exists := g.ReverseAdj[moduleID]; exists {
		return dependents
	}
	return []string{}
}

// GetModuleCount returns total number of modules
func (g *DependencyGraph) GetModuleCount() int {
	return len(g.Modules)
}

// GetDependencyCount returns total number of dependencies/edges
func (g *DependencyGraph) GetDependencyCount() int {
	return len(g.Edges)
}

// PrintGraphStatistics prints basic graph statistics
func (g *DependencyGraph) PrintGraphStatistics() {
	fmt.Println("\n=== Dependency Graph Statistics ===")
	fmt.Printf("Total Modules: %d\n", g.GetModuleCount())
	fmt.Printf("Total Dependencies: %d\n", g.GetDependencyCount())
	fmt.Printf("Average Dependencies per Module: %.2f\n", 
		float64(g.GetDependencyCount())/float64(g.GetModuleCount()))
	
	// Find highest in-degree and out-degree modules
	maxInDegree := 0
	maxOutDegree := 0
	maxInModule := ""
	maxOutModule := ""
	
	for module, inDeg := range g.InDegree {
		if inDeg > maxInDegree {
			maxInDegree = inDeg
			maxInModule = module
		}
	}
	
	for module, outDeg := range g.OutDegree {
		if outDeg > maxOutDegree {
			maxOutDegree = outDeg
			maxOutModule = module
		}
	}
	
	fmt.Printf("Most Depended-On Module: %s (in-degree: %d)\n", maxInModule, maxInDegree)
	fmt.Printf("Module with Most Dependencies: %s (out-degree: %d)\n", maxOutModule, maxOutDegree)
	fmt.Println()
}

// String returns a string representation of the graph
func (g *DependencyGraph) String() string {
	result := "DependencyGraph {\n"
	result += fmt.Sprintf("  Modules: %d\n", len(g.Modules))
	result += fmt.Sprintf("  Dependencies: %d\n", len(g.Edges))
	result += "  Edges:\n"
	
	for _, edge := range g.Edges {
		result += fmt.Sprintf("    %s → %s (weight: %d)\n", edge.Source, edge.Target, edge.Weight)
	}
	
	result += "}\n"
	return result
}
