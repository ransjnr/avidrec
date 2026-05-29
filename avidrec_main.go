package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func main() {
	// Command-line flags
	projectPath := flag.String("path", ".", "Path to the Go project to analyze")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	phase := flag.String("phase", "all", "Phase to run: all, detection, prediction")
	modelDir := flag.String("model-dir", ".avidrec_models", "Directory for trained models")
	flag.Parse()

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                        AViDRec v0.2                          ║")
	fmt.Println("║   Architectural Violation Detection and Recovery System      ║")
	fmt.Println("║         Phase 1: CycleFinder + Phase 2: ML Predictor         ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Measure execution time
	startTime := time.Now()
	var cycles []Cycle

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

	// PHASE 1: Detection
	if *phase == "all" || *phase == "detection" {
		fmt.Println("═══════════════════════════════════════════════════════════════")
		fmt.Println("PHASE 1: Architectural Violation Detection (CycleFinder)")
		fmt.Println("═══════════════════════════════════════════════════════════════")

		// Step 2: Run CycleFinder algorithm
		cycleFinder := NewCycleFinder(graph)
		cycles = cycleFinder.FindAllCycles()

		// Step 3: Print results
		cycleFinder.PrintCycles()
	}

	// PHASE 2: Prediction
	if *phase == "all" || *phase == "prediction" {
		fmt.Println("═══════════════════════════════════════════════════════════════")
		fmt.Println("PHASE 2: ML Violation Predictor (XGBoost)")
		fmt.Println("═══════════════════════════════════════════════════════════════")

		// Step 4: Generate synthetic dataset
		if cycles == nil {
			cycleFinder := NewCycleFinder(graph)
			cycles = cycleFinder.FindAllCycles()
		}

		os.MkdirAll(*modelDir, 0755)
		datasetPath := filepath.Join(*modelDir, "training_data.csv")
		
		generator := NewDatasetGenerator(graph, cycles)
		err := generator.GenerateSyntheticDataset(datasetPath)
		if err != nil {
			fmt.Printf("⚠️  Error generating dataset: %v\n", err)
		}

		// Step 5: Train ML model
		trainer := NewXGBoostTrainer(datasetPath, *modelDir)
		model, err := trainer.TrainModel()
		if err != nil {
			fmt.Printf("⚠️  Training error: %v\n", err)
		} else {
			trainer.PrintModelSummary(model)
			trainer.PrintFeatureImportance(model)
		}

		// Step 6: Make predictions on all dependencies
		if model != nil {
			predictor := NewViolationPredictor(model, graph, 0.5) // 0.5 threshold = 50%
			predictions := predictor.PredictAllDependencies(graph)
			
			// Print predictions
			predictor.PrintPredictions(predictions)

			// Print statistics
			stats := predictor.CalculatePredictionStats(predictions)
			predictor.PrintPredictionStats(stats)
		}
	}

	// Step 7: Print summary
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
	if cycles != nil {
		fmt.Printf("Circular dependencies:     %d\n", len(cycles))
	}
	fmt.Printf("Analysis time:             %.3f ms\n", elapsed.Seconds()*1000)
	fmt.Println()

	// Severity assessment
	if cycles == nil || len(cycles) == 0 {
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