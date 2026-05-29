package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
)

// DatasetGenerator generates synthetic training data for ML model
type DatasetGenerator struct {
	graph *DependencyGraph
	cycles []Cycle
	rand  *rand.Rand
}

// NewDatasetGenerator creates a new dataset generator
func NewDatasetGenerator(graph *DependencyGraph, cycles []Cycle) *DatasetGenerator {
	return &DatasetGenerator{
		graph:  graph,
		cycles: cycles,
		rand:   rand.New(rand.NewSource(42)), // Fixed seed for reproducibility
	}
}

// GenerateSyntheticDataset generates synthetic training dataset
// Uses real graph structure + synthetic features + synthetic labels based on cycles
func (dg *DatasetGenerator) GenerateSyntheticDataset(outputPath string) error {
	fmt.Printf("\n📊 Generating synthetic training dataset...\n")

	// Step 1: Extract features from real graph
	extractor := NewFeatureExtractor(dg.graph)
	allFeatures := extractor.ExtractAllFeatures()

	fmt.Printf("  Base dependencies: %d\n", len(allFeatures))

	// Step 2: Label features based on actual cycles
	dg.labelFeaturesFromCycles(allFeatures)

	// Step 3: Augment dataset with synthetic examples
	// Add variations of existing dependencies to increase dataset size
	augmentedFeatures := dg.augmentDataset(allFeatures)
	fmt.Printf("  After augmentation: %d examples\n", len(augmentedFeatures))

	// Step 4: Add synthetic negative examples
	syntheticFeatures := dg.generateSyntheticExamples(augmentedFeatures)
	fmt.Printf("  After synthetic generation: %d examples\n", len(syntheticFeatures))

	// Step 5: Write to CSV
	err := dg.writeDatasetToCSV(syntheticFeatures, outputPath)
	if err != nil {
		return err
	}

	// Print statistics
	dg.printDatasetStatistics(syntheticFeatures)

	return nil
}

// labelFeaturesFromCycles labels features based on actual cycles detected
func (dg *DatasetGenerator) labelFeaturesFromCycles(features []*DependencyFeatures) {
	// Build a map of cyclic edges
	cyclicEdges := make(map[string]map[string]bool)

	for _, cycle := range dg.cycles {
		for i := 0; i < len(cycle.Nodes)-1; i++ {
			source := cycle.Nodes[i]
			target := cycle.Nodes[i+1]

			if _, exists := cyclicEdges[source]; !exists {
				cyclicEdges[source] = make(map[string]bool)
			}
			cyclicEdges[source][target] = true
		}
	}

	// Label features
	for _, f := range features {
		if _, exists := cyclicEdges[f.SourceModule]; exists {
			if cyclicEdges[f.SourceModule][f.TargetModule] {
				f.IsViolation = 1
			}
		}
	}
}

// augmentDataset creates variations of existing dependencies
func (dg *DatasetGenerator) augmentDataset(features []*DependencyFeatures) []*DependencyFeatures {
	augmented := make([]*DependencyFeatures, 0, len(features)*3)

	// Keep original features
	for _, f := range features {
		augmented = append(augmented, f)
	}

	// Create variations (with slight feature modifications)
	for _, original := range features {
		// Variation 1: Higher complexity version
		variant1 := *original
		variant1.SourceLOC = int(float64(original.SourceLOC) * 1.5)
		variant1.TargetLOC = int(float64(original.TargetLOC) * 1.5)
		variant1.SourceCyclomatic = original.SourceCyclomatic * 1.3
		variant1.TargetCyclomatic = original.TargetCyclomatic * 1.3
		augmented = append(augmented, &variant1)

		// Variation 2: More active development
		variant2 := *original
		variant2.SourceCommits = original.SourceCommits + dg.rand.Intn(10)
		variant2.TargetCommits = original.TargetCommits + dg.rand.Intn(10)
		variant2.SourceChangeRate = original.SourceChangeRate + float64(dg.rand.Intn(20))/100.0
		variant2.TargetChangeRate = original.TargetChangeRate + float64(dg.rand.Intn(20))/100.0
		augmented = append(augmented, &variant2)
	}

	return augmented
}

// generateSyntheticExamples generates synthetic violation and non-violation examples
func (dg *DatasetGenerator) generateSyntheticExamples(baseFeatures []*DependencyFeatures) []*DependencyFeatures {
	synthetic := make([]*DependencyFeatures, 0, len(baseFeatures)*2)

	// Keep all base features
	for _, f := range baseFeatures {
		synthetic = append(synthetic, f)
	}

	// Generate synthetic violations (high-risk patterns)
	fmt.Print("  Generating synthetic violations... ")
	violations := dg.generateViolationExamples(100)
	synthetic = append(synthetic, violations...)
	fmt.Printf("%d examples\n", len(violations))

	// Generate synthetic clean dependencies (low-risk patterns)
	fmt.Print("  Generating synthetic clean dependencies... ")
	clean := dg.generateCleanExamples(100)
	synthetic = append(synthetic, clean...)
	fmt.Printf("%d examples\n", len(clean))

	return synthetic
}

// generateViolationExamples generates synthetic violation examples
func (dg *DatasetGenerator) generateViolationExamples(count int) []*DependencyFeatures {
	examples := make([]*DependencyFeatures, 0, count)

	for i := 0; i < count; i++ {
		feat := &DependencyFeatures{
			SourceModule:     fmt.Sprintf("synthetic_violation_src_%d", i),
			TargetModule:     fmt.Sprintf("synthetic_violation_tgt_%d", i),
			IsViolation:      1,
			SourceLOC:        dg.rand.Intn(2000) + 500,      // 500-2500 LOC
			TargetLOC:        dg.rand.Intn(2000) + 500,      // 500-2500 LOC
			SourceCyclomatic: float64(dg.rand.Intn(30) + 10), // High complexity
			TargetCyclomatic: float64(dg.rand.Intn(30) + 10),
			SourceCommits:    dg.rand.Intn(30) + 10,         // Active development
			TargetCommits:    dg.rand.Intn(30) + 10,
			SourceChangeRate: float64(dg.rand.Intn(100)) / 100.0,
			TargetChangeRate: float64(dg.rand.Intn(100)) / 100.0,
			SourceInDegree:   dg.rand.Intn(20) + 5, // Many dependencies
			SourceOutDegree:  dg.rand.Intn(20) + 5,
			TargetInDegree:   dg.rand.Intn(20) + 5,
			TargetOutDegree:  dg.rand.Intn(20) + 5,
			SourceTeams:      dg.rand.Intn(5) + 2,
			TargetTeams:      dg.rand.Intn(5) + 2,
			SharedTeams:      dg.rand.Intn(3) + 1, // Overlapping teams increase risk
			SourceLayer:      dg.rand.Intn(4) + 2,
			TargetLayer:      dg.rand.Intn(4) + 2,
		}
		examples = append(examples, feat)
	}

	return examples
}

// generateCleanExamples generates synthetic clean (non-violation) examples
func (dg *DatasetGenerator) generateCleanExamples(count int) []*DependencyFeatures {
	examples := make([]*DependencyFeatures, 0, count)

	for i := 0; i < count; i++ {
		feat := &DependencyFeatures{
			SourceModule:     fmt.Sprintf("synthetic_clean_src_%d", i),
			TargetModule:     fmt.Sprintf("synthetic_clean_tgt_%d", i),
			IsViolation:      0,
			SourceLOC:        dg.rand.Intn(1000) + 100,     // 100-1100 LOC
			TargetLOC:        dg.rand.Intn(1000) + 100,     // Lower complexity
			SourceCyclomatic: float64(dg.rand.Intn(10) + 1),
			TargetCyclomatic: float64(dg.rand.Intn(10) + 1),
			SourceCommits:    dg.rand.Intn(15) + 1,         // Lower change rate
			TargetCommits:    dg.rand.Intn(15) + 1,
			SourceChangeRate: float64(dg.rand.Intn(50)) / 100.0, // Lower change
			TargetChangeRate: float64(dg.rand.Intn(50)) / 100.0,
			SourceInDegree:   dg.rand.Intn(10) + 1,         // Fewer dependencies
			SourceOutDegree:  dg.rand.Intn(10) + 1,
			TargetInDegree:   dg.rand.Intn(10) + 1,
			TargetOutDegree:  dg.rand.Intn(10) + 1,
			SourceTeams:      dg.rand.Intn(3) + 1,
			TargetTeams:      dg.rand.Intn(3) + 1,
			SharedTeams:      0, // No shared teams = clean boundary
			SourceLayer:      dg.rand.Intn(2),              // Lower layers (foundation)
			TargetLayer:      dg.rand.Intn(2),
		}
		examples = append(examples, feat)
	}

	return examples
}

// writeDatasetToCSV writes features to CSV file for training
func (dg *DatasetGenerator) writeDatasetToCSV(features []*DependencyFeatures, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	featureNames := GetFeatureNames()
	header := append(featureNames, "IsViolation")
	writer.Write(header)

	// Write data rows
	for _, f := range features {
		values := FeaturesToArray(f)
		row := make([]string, 0, len(values)+1)
		for _, v := range values {
			row = append(row, fmt.Sprintf("%.4f", v))
		}
		row = append(row, strconv.Itoa(f.IsViolation))
		writer.Write(row)
	}

	fmt.Printf("✓ Dataset saved to: %s\n", outputPath)
	return nil
}

// printDatasetStatistics prints statistics about the dataset
func (dg *DatasetGenerator) printDatasetStatistics(features []*DependencyFeatures) {
	violationCount := 0
	cleanCount := 0

	avgSourceLOC := 0.0
	avgTargetLOC := 0.0
	avgSourceCommits := 0.0
	avgTargetCommits := 0.0

	for _, f := range features {
		if f.IsViolation == 1 {
			violationCount++
		} else {
			cleanCount++
		}

		avgSourceLOC += float64(f.SourceLOC)
		avgTargetLOC += float64(f.TargetLOC)
		avgSourceCommits += float64(f.SourceCommits)
		avgTargetCommits += float64(f.TargetCommits)
	}

	n := float64(len(features))
	avgSourceLOC /= n
	avgTargetLOC /= n
	avgSourceCommits /= n
	avgTargetCommits /= n

	fmt.Println("\n=== Dataset Statistics ===")
	fmt.Printf("Total examples: %d\n", len(features))
	fmt.Printf("Violations (label=1): %d (%.1f%%)\n", violationCount, float64(violationCount)*100/n)
	fmt.Printf("Clean (label=0): %d (%.1f%%)\n", cleanCount, float64(cleanCount)*100/n)
	fmt.Printf("Class balance ratio: %.2f:1\n", float64(violationCount)/float64(cleanCount))
	fmt.Println()
	fmt.Printf("Average Source LOC: %.0f\n", avgSourceLOC)
	fmt.Printf("Average Target LOC: %.0f\n", avgTargetLOC)
	fmt.Printf("Average Source Commits: %.1f\n", avgSourceCommits)
	fmt.Printf("Average Target Commits: %.1f\n", avgTargetCommits)
	fmt.Println()
}

// FeatureCorrelation calculates correlation between a feature and violation label
func (dg *DatasetGenerator) FeatureCorrelation(features []*DependencyFeatures, featureIndex int) float64 {
	if len(features) == 0 {
		return 0
	}

	// Extract feature values and labels
	featureValues := make([]float64, len(features))
	labels := make([]float64, len(features))

	for i, f := range features {
		featureValues[i] = FeaturesToArray(f)[featureIndex]
		labels[i] = float64(f.IsViolation)
	}

	// Calculate mean
	meanFeature := 0.0
	meanLabel := 0.0
	for i := range features {
		meanFeature += featureValues[i]
		meanLabel += labels[i]
	}
	meanFeature /= float64(len(features))
	meanLabel /= float64(len(features))

	// Calculate correlation
	numerator := 0.0
	denomFeature := 0.0
	denomLabel := 0.0

	for i := range features {
		diffFeature := featureValues[i] - meanFeature
		diffLabel := labels[i] - meanLabel
		numerator += diffFeature * diffLabel
		denomFeature += diffFeature * diffFeature
		denomLabel += diffLabel * diffLabel
	}

	if denomFeature == 0 || denomLabel == 0 {
		return 0
	}

	correlation := numerator / math.Sqrt(denomFeature*denomLabel)
	return correlation
}