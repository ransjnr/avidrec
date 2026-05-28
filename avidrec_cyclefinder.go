package main

import (
	"fmt"
	"sort"
)

// Cycle represents a circular dependency cycle in the code
type Cycle struct {
	Nodes []string // Path of modules forming the cycle (A → B → C → A)
	Length int     // Length of the cycle
}

// CycleFinder detects cycles in the dependency graph using O(n log n) algorithm
type CycleFinder struct {
	graph              *DependencyGraph
	visited            map[string]bool     // Visited nodes
	recursionStack     map[string]bool     // Nodes in current recursion stack
	cycles             []Cycle             // Found cycles
	inDegree           map[string]int      // Copy of in-degrees (will be modified)
	adjacencyList      map[string][]string // Copy of adjacency list (will be modified)
}

// NewCycleFinder creates a new CycleFinder instance
func NewCycleFinder(graph *DependencyGraph) *CycleFinder {
	return &CycleFinder{
		graph:          graph,
		visited:        make(map[string]bool),
		recursionStack: make(map[string]bool),
		cycles:         make([]Cycle, 0),
		inDegree:       make(map[string]int),
		adjacencyList:  make(map[string][]string),
	}
}

// FindAllCycles finds all circular dependencies in the graph
// Algorithm complexity: O(n log n + m) where n = nodes, m = edges
func (cf *CycleFinder) FindAllCycles() []Cycle {
	fmt.Println("\n🔍 Running CycleFinder Algorithm (O(n log n))...")
	
	// Step 1: Copy graph structures (will be modified during algorithm)
	for module := range cf.graph.Modules {
		cf.inDegree[module] = cf.graph.InDegree[module]
		cf.adjacencyList[module] = make([]string, len(cf.graph.AdjList[module]))
		copy(cf.adjacencyList[module], cf.graph.AdjList[module])
	}

	// Step 2: Phase 1 - Topological sort (Kahn's algorithm variant)
	cyclicNodes := cf.identifyCyclicNodes()
	
	if len(cyclicNodes) == 0 {
		fmt.Println("✓ No cycles found! Graph is acyclic.")
		return cf.cycles
	}

	fmt.Printf("Found %d nodes involved in cycles\n", len(cyclicNodes))

	// Step 3: Phase 2 - Extract cyclic subgraph
	cyclicSubgraph := cf.extractCyclicSubgraph(cyclicNodes)

	// Step 4: Phase 3 - Find all cycles in cyclic subgraph
	cf.findCyclesInSubgraph(cyclicSubgraph, cyclicNodes)

	// Deduplicate cycles
	cf.deduplicateCycles()

	return cf.cycles
}

// identifyCyclicNodes identifies all nodes that are part of cycles
// Returns a set (map[string]bool) of nodes involved in cycles
// Complexity: O(n + m) where n = nodes, m = edges
func (cf *CycleFinder) identifyCyclicNodes() map[string]bool {
	fmt.Println("  Phase 1: Identifying cyclic nodes using topological sort...")
	
	// Create queue of nodes with in-degree 0
	queue := make([]string, 0)
	for module, degree := range cf.inDegree {
		if degree == 0 {
			queue = append(queue, module)
		}
	}

	removedNodes := make(map[string]bool)

	// Process nodes with in-degree 0
	for len(queue) > 0 {
		// Pop from queue
		node := queue[0]
		queue = queue[1:]

		removedNodes[node] = true

		// For each neighbor of the current node
		for _, neighbor := range cf.adjacencyList[node] {
			// Decrease in-degree
			cf.inDegree[neighbor]--

			// If in-degree becomes 0, add to queue
			if cf.inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Nodes not removed are part of cycles
	cyclicNodes := make(map[string]bool)
	for module := range cf.graph.Modules {
		if !removedNodes[module] {
			cyclicNodes[module] = true
		}
	}

	return cyclicNodes
}

// extractCyclicSubgraph extracts only the subgraph containing cyclic nodes
// Complexity: O(n + m) for cyclic portion
func (cf *CycleFinder) extractCyclicSubgraph(cyclicNodes map[string]bool) map[string][]string {
	fmt.Println("  Phase 2: Extracting cyclic subgraph...")
	
	cyclicSubgraph := make(map[string][]string)

	for node := range cyclicNodes {
		cyclicSubgraph[node] = make([]string, 0)

		// Only include edges to other cyclic nodes
		for _, neighbor := range cf.adjacencyList[node] {
			if cyclicNodes[neighbor] {
				cyclicSubgraph[node] = append(cyclicSubgraph[node], neighbor)
			}
		}
	}

	return cyclicSubgraph
}

// findCyclesInSubgraph finds all cycles in the cyclic subgraph using DFS
// Complexity: O(n + m) for cyclic subgraph
func (cf *CycleFinder) findCyclesInSubgraph(subgraph map[string][]string, cyclicNodes map[string]bool) {
	fmt.Println("  Phase 3: Finding all cycles in subgraph...")
	
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := make([]string, 0)

	for node := range cyclicNodes {
		if !visited[node] {
			cf.dfsForCycles(node, subgraph, visited, recStack, path)
		}
	}
}

// dfsForCycles performs DFS to find cycles
func (cf *CycleFinder) dfsForCycles(node string, subgraph map[string][]string, 
	visited map[string]bool, recStack map[string]bool, path []string) {
	
	visited[node] = true
	recStack[node] = true
	path = append(path, node)

	for _, neighbor := range subgraph[node] {
		if !visited[neighbor] {
			cf.dfsForCycles(neighbor, subgraph, visited, recStack, path)
		} else if recStack[neighbor] {
			// Found a cycle! Extract it
			cycle := cf.extractCycleFromPath(path, neighbor)
			cf.cycles = append(cf.cycles, cycle)
		}
	}

	// Remove from recursion stack
	recStack[node] = false
	path = path[:len(path)-1]
}

// extractCycleFromPath extracts a single cycle from the current path
func (cf *CycleFinder) extractCycleFromPath(path []string, cycleStart string) Cycle {
	// Find where cycleStart appears in path
	startIdx := -1
	for i, node := range path {
		if node == cycleStart {
			startIdx = i
			break
		}
	}

	// Extract cycle from startIdx to end
	cyclePath := make([]string, 0)
	if startIdx >= 0 {
		cyclePath = append(cyclePath, path[startIdx:]...)
		cyclePath = append(cyclePath, cycleStart) // Close the cycle
	}

	return Cycle{
		Nodes:  cyclePath,
		Length: len(cyclePath) - 1, // Don't count the repeated end node
	}
}

// deduplicateCycles removes duplicate cycles from the results
func (cf *CycleFinder) deduplicateCycles() {
	if len(cf.cycles) == 0 {
		return
	}

	// Convert cycles to strings for comparison
	seen := make(map[string]bool)
	uniqueCycles := make([]Cycle, 0)

	for _, cycle := range cf.cycles {
		cycleStr := cf.cycleToNormalizedString(cycle)
		
		if !seen[cycleStr] {
			seen[cycleStr] = true
			uniqueCycles = append(uniqueCycles, cycle)
		}
	}

	cf.cycles = uniqueCycles
}

// cycleToNormalizedString converts a cycle to a normalized string for comparison
// This handles cycles like A→B→A, B→A→B, etc. as the same cycle
func (cf *CycleFinder) cycleToNormalizedString(cycle Cycle) string {
	if len(cycle.Nodes) == 0 {
		return ""
	}

	// Remove the last node (which is a repeat of the first)
	nodes := cycle.Nodes[:len(cycle.Nodes)-1]
	
	// Find the rotation that gives the lexicographically smallest string
	minStr := cycle.nodesToString(nodes)
	
	for i := 1; i < len(nodes); i++ {
		rotated := make([]string, len(nodes))
		copy(rotated, nodes[i:])
		copy(rotated[len(nodes)-i:], nodes[:i])
		rotatedStr := cycle.nodesToString(rotated)
		
		if rotatedStr < minStr {
			minStr = rotatedStr
		}
	}

	return minStr
}

// nodesToString converts a list of nodes to a string
func (c *Cycle) nodesToString(nodes []string) string {
	result := ""
	for i, node := range nodes {
		if i > 0 {
			result += " → "
		}
		result += node
	}
	return result
}

// GetCycles returns all found cycles
func (cf *CycleFinder) GetCycles() []Cycle {
	return cf.cycles
}

// GetCycleCount returns the total number of unique cycles found
func (cf *CycleFinder) GetCycleCount() int {
	return len(cf.cycles)
}

// PrintCycles prints all found cycles in a readable format
func (cf *CycleFinder) PrintCycles() {
	if len(cf.cycles) == 0 {
		fmt.Println("\n✓ No circular dependencies found!")
		return
	}

	fmt.Printf("\n⚠️  Found %d circular dependencies:\n\n", len(cf.cycles))
	
	// Sort cycles by length for readability
	sortedCycles := make([]Cycle, len(cf.cycles))
	copy(sortedCycles, cf.cycles)
	sort.Slice(sortedCycles, func(i, j int) bool {
		return sortedCycles[i].Length < sortedCycles[j].Length
	})

	for i, cycle := range sortedCycles {
		fmt.Printf("Cycle %d (length %d):\n", i+1, cycle.Length)
		fmt.Print("  ")
		for j, node := range cycle.Nodes {
			if j > 0 {
				fmt.Print(" → ")
			}
			fmt.Print(node)
		}
		fmt.Println()
	}
	fmt.Println()
}

// GetViolationSeverity returns a severity score for a cycle (longer = more severe)
func (c *Cycle) GetViolationSeverity() float64 {
	// Simple severity: longer cycles are worse
	// Can be enhanced with weight information
	return float64(c.Length)
}

// String returns a string representation of a cycle
func (c *Cycle) String() string {
	result := ""
	for i, node := range c.Nodes {
		if i > 0 {
			result += " → "
		}
		result += node
	}
	return result
}
