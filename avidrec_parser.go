package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// GoCodeParser parses Go source files and extracts module dependencies
type GoCodeParser struct {
	ProjectRoot string            // Root directory of the project
	ModuleName  string            // Module name (from go.mod)
	Modules     map[string]bool   // Set of discovered modules
	Dependencies map[string][]string // Module -> list of imports
}

// NewGoCodeParser creates a new Go code parser
func NewGoCodeParser(projectRoot string) *GoCodeParser {
	return &GoCodeParser{
		ProjectRoot:  projectRoot,
		Modules:      make(map[string]bool),
		Dependencies: make(map[string][]string),
	}
}

// ParseProject walks through the project directory and extracts all dependencies
func (p *GoCodeParser) ParseProject() (*DependencyGraph, error) {
	absRoot, err := filepath.Abs(p.ProjectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project path: %v", err)
	}
	p.ProjectRoot = absRoot

	// Step 1: Read go.mod to get module name
	err = p.readGoMod()
	if err != nil {
		return nil, fmt.Errorf("failed to read go.mod: %v", err)
	}

	fmt.Printf("Parsing Go project: %s\n", p.ModuleName)

	// Step 2: Walk through all Go files
	err = filepath.Walk(p.ProjectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process .go files
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			// Skip test files for now (can add later)
			if strings.HasSuffix(path, "_test.go") {
				return nil
			}

			relPath, err := filepath.Rel(p.ProjectRoot, path)
			if err != nil {
				return err
			}
			pkg := p.getPackagePath(relPath)

			if err = p.parseGoFile(path, pkg); err != nil {
				fmt.Printf("Warning: Failed to parse %s: %v\n", path, err)
				// Continue parsing other files
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk project: %v", err)
	}

	// Step 3: Build and return the dependency graph
	graph := p.buildGraph()
	return graph, nil
}

// readGoMod reads the go.mod file to extract module name
func (p *GoCodeParser) readGoMod() error {
	goModPath := filepath.Join(p.ProjectRoot, "go.mod")
	
	data, err := ioutil.ReadFile(goModPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "module ") {
			p.ModuleName = strings.TrimPrefix(line, "module ")
			p.ModuleName = strings.TrimSpace(p.ModuleName)
			return nil
		}
	}

	return fmt.Errorf("could not find module directive in go.mod")
}

// getPackagePath converts a file path to a package path
func (p *GoCodeParser) getPackagePath(filePath string) string {
	dir := filepath.Dir(filePath)
	dir = strings.ReplaceAll(dir, string(os.PathSeparator), "/")
	
	if dir == "." || dir == "" {
		return p.ModuleName
	}
	
	return p.ModuleName + "/" + dir
}

// parseGoFile parses a single Go file and extracts imports
func (p *GoCodeParser) parseGoFile(filePath string, packagePath string) error {
	// Read and parse the Go file using Go's AST parser
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ImportsOnly)
	if err != nil {
		return fmt.Errorf("failed to parse file: %v", err)
	}

	// Register this module/package
	p.Modules[packagePath] = true
	if _, exists := p.Dependencies[packagePath]; !exists {
		p.Dependencies[packagePath] = make([]string, 0)
	}

	// Extract imports from the file
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")
		
		// Only track internal imports (those within our module)
		if strings.HasPrefix(importPath, p.ModuleName) {
			p.Dependencies[packagePath] = append(p.Dependencies[packagePath], importPath)
		}
	}

	return nil
}

// buildGraph converts parsed dependencies into a DependencyGraph
func (p *GoCodeParser) buildGraph() *DependencyGraph {
	graph := NewDependencyGraph()

	// Add all modules as nodes
	for module := range p.Modules {
		m := &Module{
			ID:   module,
			Name: p.getModuleName(module),
		}
		graph.AddModule(m)
	}

	// Add all dependencies as edges
	for source, targets := range p.Dependencies {
		for _, target := range targets {
			err := graph.AddDependency(source, target, 1)
			if err != nil {
				fmt.Printf("Warning: %v\n", err)
			}
		}
	}

	return graph
}

// getModuleName extracts the short module name from a full path
func (p *GoCodeParser) getModuleName(modulePath string) string {
	parts := strings.Split(modulePath, "/")
	return parts[len(parts)-1]
}

// PrintDiscoveredModules prints all discovered modules
func (p *GoCodeParser) PrintDiscoveredModules() {
	fmt.Println("\n=== Discovered Modules ===")
	for module := range p.Modules {
		fmt.Printf("  %s\n", module)
	}
	fmt.Printf("Total: %d modules\n\n", len(p.Modules))
}

// PrintDependencies prints all discovered dependencies
func (p *GoCodeParser) PrintDependencies() {
	fmt.Println("\n=== Module Dependencies ===")
	for module, deps := range p.Dependencies {
		if len(deps) > 0 {
			fmt.Printf("%s depends on:\n", module)
			for _, dep := range deps {
				fmt.Printf("  → %s\n", dep)
			}
		}
	}
	fmt.Println()
}
