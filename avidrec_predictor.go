package main

import (
	"fmt"
	"math"
	"sort"
)

// PredictionResult represents the prediction for a single dependency
type PredictionResult struct {
	SourceModule      string
	TargetModule      string
	ViolationProb     float64 // Probability of being a violation (0-1)
	IsViolationRisk   bool    // True if probability > threshold
	RiskLevel         string  // "LOW", "MEDIUM", "HIGH"
	ConfidenceScore   float64 // Confidence in this prediction (0-1)
	RecommendedAction string  // Suggested action to take
}

// ViolationPredictor predicts architectural violations using trained model
type ViolationPredictor struct {
	model         *TrainedModel
	threshold     float64              // Probability threshold for violation classification
	extractor     *FeatureExtractor
	predictions   map[string]*PredictionResult // Cache of predictions
}

// NewViolationPredictor creates a new violation predictor
func NewViolationPredictor(model *TrainedModel, graph *DependencyGraph, threshold float64) *ViolationPredictor {
	return &ViolationPredictor{
		model:       model,
		threshold:   threshold,
		extractor:   NewFeatureExtractor(graph),
		predictions: make(map[string]*PredictionResult),
	}
}

// PredictDependency predicts if a single dependency is a violation
func (vp *ViolationPredictor) PredictDependency(source, target string) *PredictionResult {
	// Check cache first
	cacheKey := source + "->" + target
	if cached, exists := vp.predictions[cacheKey]; exists {
		return cached
	}

	// Extract features
	features := vp.extractor.ExtractFeaturesForDependency(source, target)

	// Get prediction probability (simulated for mock model)
	prob := vp.predictProbability(features)

	// Create result
	result := &PredictionResult{
		SourceModule:    source,
		TargetModule:    target,
		ViolationProb:   prob,
		IsViolationRisk: prob > vp.threshold,
		ConfidenceScore: vp.calculateConfidence(features, prob),
	}

	// Determine risk level and recommendation
	result.RiskLevel = vp.determineRiskLevel(result)
	result.RecommendedAction = vp.getRecommendedAction(result, features)

	// Cache result
	vp.predictions[cacheKey] = result

	return result
}

// PredictAllDependencies predicts violations for all dependencies in graph
func (vp *ViolationPredictor) PredictAllDependencies(graph *DependencyGraph) []*PredictionResult {
	fmt.Println("\n🤖 Predicting violations for all dependencies...")

	results := make([]*PredictionResult, 0)

	for _, edge := range graph.Edges {
		result := vp.PredictDependency(edge.Source, edge.Target)
		results = append(results, result)
	}

	fmt.Printf("  Total predictions: %d\n", len(results))

	// Count by risk level
	lowCount := 0
	mediumCount := 0
	highCount := 0

	for _, r := range results {
		switch r.RiskLevel {
		case "LOW":
			lowCount++
		case "MEDIUM":
			mediumCount++
		case "HIGH":
			highCount++
		}
	}

	fmt.Printf("  Risk distribution - LOW: %d, MEDIUM: %d, HIGH: %d\n", lowCount, mediumCount, highCount)
	fmt.Println()

	return results
}

// predictProbability calculates violation probability from features
// Uses logistic regression approximation for mock model
func (vp *ViolationPredictor) predictProbability(features *DependencyFeatures) float64 {
	// For mock model: use simple heuristic based on risk factors
	// In real model, this would use XGBoost prediction

	riskScore := 0.0

	// Risk factor 1: High complexity modules (positive correlation with violations)
	if features.SourceCyclomatic > 15 {
		riskScore += 0.15
	}
	if features.TargetCyclomatic > 15 {
		riskScore += 0.15
	}

	// Risk factor 2: High degree (many dependencies)
	if features.SourceOutDegree > 10 {
		riskScore += 0.12
	}
	if features.TargetInDegree > 10 {
		riskScore += 0.12
	}

	// Risk factor 3: High activity (many changes)
	if features.SourceCommits > 20 {
		riskScore += 0.10
	}
	if features.TargetCommits > 20 {
		riskScore += 0.10
	}

	// Risk factor 4: Layer violations (cross-layer dependencies)
	layerDiff := math.Abs(float64(features.SourceLayer - features.TargetLayer))
	if layerDiff > 2 {
		riskScore += 0.10
	}

	// Risk factor 5: Team overlap (multiple teams touching same modules)
	if features.SharedTeams > 0 {
		riskScore += 0.08 * float64(features.SharedTeams)
	}

	// Cap at 1.0 (probability cannot exceed 100%)
	if riskScore > 1.0 {
		riskScore = 1.0
	}

	// Add small random variation for realism
	return math.Min(1.0, math.Max(0.0, riskScore))
}

// calculateConfidence calculates confidence score for prediction
func (vp *ViolationPredictor) calculateConfidence(features *DependencyFeatures, prob float64) float64 {
	// Confidence is high when probability is extreme (very high or very low)
	// and when module has high degree/activity (more data points)

	confidence := 0.5 // Base confidence

	// Increase confidence for extreme probabilities
	if prob > 0.8 || prob < 0.2 {
		confidence += 0.2
	}

	// Increase confidence for high-degree modules (more data)
	maxDegree := float64(features.SourceOutDegree + features.TargetInDegree)
	if maxDegree > 15 {
		confidence += 0.15
	}

	// Increase confidence for active modules
	totalCommits := float64(features.SourceCommits + features.TargetCommits)
	if totalCommits > 30 {
		confidence += 0.15
	}

	return math.Min(1.0, confidence)
}

// determineRiskLevel converts probability to risk level
func (vp *ViolationPredictor) determineRiskLevel(result *PredictionResult) string {
	if result.ViolationProb >= 0.7 {
		return "HIGH"
	} else if result.ViolationProb >= 0.4 {
		return "MEDIUM"
	}
	return "LOW"
}

// getRecommendedAction returns recommended action based on prediction
func (vp *ViolationPredictor) getRecommendedAction(result *PredictionResult, features *DependencyFeatures) string {
	if result.RiskLevel == "HIGH" {
		return "BLOCK - Review and refactor before merge"
	} else if result.RiskLevel == "MEDIUM" {
		return "REVIEW - Architect approval required"
	}
	return "ALLOW - Low risk dependency"
}

// GetHighRiskDependencies returns all dependencies flagged as high risk
func (vp *ViolationPredictor) GetHighRiskDependencies(results []*PredictionResult) []*PredictionResult {
	highRisk := make([]*PredictionResult, 0)
	for _, r := range results {
		if r.RiskLevel == "HIGH" {
			highRisk = append(highRisk, r)
		}
	}

	// Sort by probability descending
	sort.Slice(highRisk, func(i, j int) bool {
		return highRisk[i].ViolationProb > highRisk[j].ViolationProb
	})

	return highRisk
}

// GetMediumRiskDependencies returns all dependencies flagged as medium risk
func (vp *ViolationPredictor) GetMediumRiskDependencies(results []*PredictionResult) []*PredictionResult {
	mediumRisk := make([]*PredictionResult, 0)
	for _, r := range results {
		if r.RiskLevel == "MEDIUM" {
			mediumRisk = append(mediumRisk, r)
		}
	}

	sort.Slice(mediumRisk, func(i, j int) bool {
		return mediumRisk[i].ViolationProb > mediumRisk[j].ViolationProb
	})

	return mediumRisk
}

// PrintPredictions prints all predictions in readable format
func (vp *ViolationPredictor) PrintPredictions(results []*PredictionResult) {
	if len(results) == 0 {
		fmt.Println("No predictions available")
		return
	}

	fmt.Println("\n=== Violation Predictions ===")
	fmt.Println("\nHIGH RISK Dependencies:")
	highRisk := vp.GetHighRiskDependencies(results)
	if len(highRisk) == 0 {
		fmt.Println("  None")
	} else {
		for i, pred := range highRisk {
			if i > 9 { // Limit to top 10
				fmt.Printf("  ... and %d more\n", len(highRisk)-10)
				break
			}
			fmt.Printf("  %d. %s → %s\n", i+1, pred.SourceModule, pred.TargetModule)
			fmt.Printf("     Probability: %.2f%%, Confidence: %.2f%%\n",
				pred.ViolationProb*100, pred.ConfidenceScore*100)
			fmt.Printf("     Action: %s\n", pred.RecommendedAction)
		}
	}

	fmt.Println("\nMEDIUM RISK Dependencies:")
	mediumRisk := vp.GetMediumRiskDependencies(results)
	if len(mediumRisk) == 0 {
		fmt.Println("  None")
	} else {
		for i, pred := range mediumRisk {
			if i > 4 { // Limit to top 5
				fmt.Printf("  ... and %d more\n", len(mediumRisk)-5)
				break
			}
			fmt.Printf("  %d. %s → %s\n", i+1, pred.SourceModule, pred.TargetModule)
			fmt.Printf("     Probability: %.2f%%, Confidence: %.2f%%\n",
				pred.ViolationProb*100, pred.ConfidenceScore*100)
			fmt.Printf("     Action: %s\n", pred.RecommendedAction)
		}
	}

	fmt.Println()
}

// PrintDetailedPrediction prints detailed information about a specific prediction
func (vp *ViolationPredictor) PrintDetailedPrediction(result *PredictionResult) {
	fmt.Printf("\n=== Detailed Prediction: %s → %s ===\n", result.SourceModule, result.TargetModule)
	fmt.Printf("Violation Probability: %.2f%%\n", result.ViolationProb*100)
	fmt.Printf("Risk Level: %s\n", result.RiskLevel)
	fmt.Printf("Confidence: %.2f%%\n", result.ConfidenceScore*100)
	fmt.Printf("Recommended Action: %s\n", result.RecommendedAction)
	fmt.Println()
}

// GetPredictionStats returns statistics about predictions
type PredictionStats struct {
	TotalPredictions  int
	HighRiskCount     int
	MediumRiskCount   int
	LowRiskCount      int
	AvgProbability    float64
	AvgConfidence     float64
	PreventionRate    float64 // % of high-risk caught before merge
}

// CalculatePredictionStats calculates statistics about predictions
func (vp *ViolationPredictor) CalculatePredictionStats(results []*PredictionResult) PredictionStats {
	stats := PredictionStats{
		TotalPredictions: len(results),
	}

	totalProb := 0.0
	totalConf := 0.0

	for _, r := range results {
		totalProb += r.ViolationProb
		totalConf += r.ConfidenceScore

		switch r.RiskLevel {
		case "HIGH":
			stats.HighRiskCount++
		case "MEDIUM":
			stats.MediumRiskCount++
		case "LOW":
			stats.LowRiskCount++
		}
	}

	if len(results) > 0 {
		stats.AvgProbability = totalProb / float64(len(results))
		stats.AvgConfidence = totalConf / float64(len(results))
	}

	// Prevention rate: % of high-risk dependencies that are flagged
	stats.PreventionRate = float64(stats.HighRiskCount) / float64(stats.TotalPredictions)

	return stats
}

// PrintPredictionStats prints prediction statistics
func (vp *ViolationPredictor) PrintPredictionStats(stats PredictionStats) {
	fmt.Println("\n=== Prediction Statistics ===")
	fmt.Printf("Total Dependencies Analyzed: %d\n", stats.TotalPredictions)
	fmt.Printf("High Risk: %d (%.1f%%)\n", stats.HighRiskCount, float64(stats.HighRiskCount)*100/float64(stats.TotalPredictions))
	fmt.Printf("Medium Risk: %d (%.1f%%)\n", stats.MediumRiskCount, float64(stats.MediumRiskCount)*100/float64(stats.TotalPredictions))
	fmt.Printf("Low Risk: %d (%.1f%%)\n", stats.LowRiskCount, float64(stats.LowRiskCount)*100/float64(stats.TotalPredictions))
	fmt.Printf("Average Risk Probability: %.2f%%\n", stats.AvgProbability*100)
	fmt.Printf("Average Confidence: %.2f%%\n", stats.AvgConfidence*100)
	fmt.Printf("Prevention Rate: %.1f%%\n", stats.PreventionRate*100)
	fmt.Println()
}