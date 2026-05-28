package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	// Command-line flags
	projectPath := flag.String("path", ".", "Path to the Go project to analyze")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	flag.Parse()

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                        AViDRec v0.1                          ║")
	fmt.Println("║   Architectural Violation Detection and Recovery System      ║")
	fmt.Println("║                 CycleFinder Algorithm Demo                   ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Measure execution time
	startTime := time.Now()

	// Step 1: Parse the Go project
	fmt.Printf("📂 Analyzing project at: %s\n", *projectPath)
	parser := NewGoCodeParser(*projectPath)
	
	graph, err := parser.ParseProject()
	if err != nil {
		fmt.Printf("❌ Error parsing project: %v\n", err)
		os.Exit(1)
	}

	// Print discovered modules if verbose
	if *verbose {
		parser.PrintDiscoveredModules()
		parser.PrintDependencies()
	}

	// Print graph statistics
	graph.PrintGraphStatistics()

	// Step 2: Run CycleFinder algorithm
	cycleFinder := NewCycleFinder(graph)
	cycles := cycleFinder.FindAllCycles()

	// Step 3: Print results
	cycleFinder.PrintCycles()

	// Step 4: Print summary
	elapsed := time.Since(startTime)
	printSummary(graph, cycles, elapsed)
}

// printSummary prints a summary report of the analysis
func printSummary(graph *DependencyGraph, cycles []Cycle, elapsed time.Duration) {
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("                       📊 ANALYSIS SUMMARY")
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Printf("Modules analyzed:          %d\n", graph.GetModuleCount())
	fmt.Printf("Dependencies found:        %d\n", graph.GetDependencyCount())
	fmt.Printf("Circular dependencies:     %d\n", len(cycles))
	fmt.Printf("Analysis time:             %.3f ms\n", elapsed.Seconds()*1000)
	fmt.Println()

	// Severity assessment
	if len(cycles) == 0 {
		fmt.Println("✓ RESULT: No architectural violations detected!")
		fmt.Println("  The codebase has a healthy dependency structure.")
	} else {
		fmt.Printf("⚠️  RESULT: %d architectural violations detected!\n", len(cycles))
		fmt.Println("  These circular dependencies should be resolved to improve:")
		fmt.Println("    - Code maintainability")
		fmt.Println("    - Build times")
		fmt.Println("    - System stability")
		fmt.Println("    - Team productivity")
	}

	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println()
}
