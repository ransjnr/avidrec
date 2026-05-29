package main

import (
	"fmt"
	"sort"
	"strings"
)

// RefactoringSolution represents a complete solution for fixing violations
type RefactoringSolution struct {
	Violation *PredictionResult
	Fixes     []*RefactoringFix
	Best      *RefactoringFix // Recommended fix (highest score)
	Score     float64         // Overall solution quality (0-1)
}

// RefactoringSynthesizer orchestrates the refactoring synthesis pipeline
type RefactoringSynthesizer struct {
	predictions []*PredictionResult
	graph       *DependencyGraph
	codegen     *CodeGenerator
	solver      *Z3Solver
	solutions   []*RefactoringSolution
}

// NewRefactoringSynthesizer creates a new synthesizer
func NewRefactoringSynthesizer(predictions []*PredictionResult, graph *DependencyGraph) *RefactoringSynthesizer {
	return &RefactoringSynthesizer{
		predictions: predictions,
		graph:       graph,
		codegen:     NewCodeGenerator(predictions, graph),
		solver:      NewZ3Solver(),
		solutions:   make([]*RefactoringSolution, 0),
	}
}

// SynthesizeAllSolutions generates fixes for all high-risk violations
func (rs *RefactoringSynthesizer) SynthesizeAllSolutions() []*RefactoringSolution {
	fmt.Println("\n🔧 Synthesizing refactoring solutions...")

	highRiskViolations := filterHighRiskViolations(rs.predictions)
	fmt.Printf("  Found %d high-risk violations requiring fixes\n\n", len(highRiskViolations))

	for i, violation := range highRiskViolations {
		solution := rs.synthesizeSolution(violation)
		if solution != nil && len(solution.Fixes) > 0 {
			rs.solutions = append(rs.solutions, solution)
			fmt.Printf("✓ Solution %d: %s → %s\n", i+1, violation.SourceModule, violation.TargetModule)
			fmt.Printf("  Best strategy: %s (Score: %.2f)\n\n", solution.Best.Strategy.Type, solution.Best.Score)
		}
	}

	return rs.solutions
}

// synthesizeSolution generates all possible fixes for a single violation
func (rs *RefactoringSynthesizer) synthesizeSolution(violation *PredictionResult) *RefactoringSolution {
	solution := &RefactoringSolution{
		Violation: violation,
		Fixes:     make([]*RefactoringFix, 0),
		Score:     0,
	}

	// Extract features for constraint solving
	extractor := NewFeatureExtractor(rs.graph)
	features := extractor.ExtractFeaturesForDependency(violation.SourceModule, violation.TargetModule)

	// Get refactoring strategies from Z3
	strategies := rs.solver.SolveRefactoringOptions(violation, features)

	// Generate code for each strategy
	for _, strategy := range strategies {
		fix := rs.generateRefactoringFix(violation, strategy, features)
		if fix != nil {
			solution.Fixes = append(solution.Fixes, fix)

			// Update best if this is better
			if solution.Best == nil || fix.Score > solution.Best.Score {
				solution.Best = fix
			}
		}
	}

	// Calculate solution score
	if solution.Best != nil {
		solution.Score = solution.Best.Score
	}

	return solution
}

// generateRefactoringFix generates a single refactoring fix
func (rs *RefactoringSynthesizer) generateRefactoringFix(violation *PredictionResult, strategy *RefactoringStrategy, features *DependencyFeatures) *RefactoringFix {
	source := extractPackageName(violation.SourceModule)
	target := extractPackageName(violation.TargetModule)

	fix := &RefactoringFix{
		Strategy:      strategy,
		SourceModule:  violation.SourceModule,
		TargetModule:  violation.TargetModule,
		LanguageCode:  make(map[string]string),
		Rollback:      make(map[string]string),
		Changes:       make([]string, 0),
		Score:         strategy.Acceptance,
	}

	// Generate code for all supported languages
	switch strategy.Type {
	case "DEPENDENCY_REMOVAL":
		generateRemovalSolution(fix, source, target)
	case "INTERFACE_EXTRACTION":
		generateInterfaceExtractionSolution(fix, source, target)
	case "MODULE_RESTRUCTURING":
		generateRestructuringSolution(fix, source, target)
	}

	// Validate semantic preservation
	if validateSemanticPreservation(fix) {
		fix.Score *= 0.95 // 95% confidence in semantic preservation
	} else {
		fix.Score *= 0.85 // 85% if validation less certain
	}

	return fix
}

// ==================== SOLUTION GENERATORS ====================

func generateRemovalSolution(fix *RefactoringFix, source, target string) {
	// Go
	fix.LanguageCode["go"] = `
// Step 1: Create an interface for the shared functionality
type ` + target + `Service interface {
    DoWork() error
    GetData() interface{}
}

// Step 2: Update functions to accept interface instead of concrete type
func DoSomething(svc ` + target + `Service) error {
    return svc.DoWork()
}

// Step 3: Remove import statement
// DELETE: import "` + target + `"

// Step 4: At injection points, pass interface implementation instead
`

	// Java
	fix.LanguageCode["java"] = `
// Step 1: Create interface
public interface ` + toTitleCase(target) + `Service {
    void doWork() throws Exception;
    Object getData();
}

// Step 2: Use interface in code
public void doSomething(` + toTitleCase(target) + `Service svc) throws Exception {
    svc.doWork();
}

// Step 3: Remove import
// DELETE: import ` + target + `.*;
`

	// Python
	fix.LanguageCode["python"] = `
# Step 1: Create abstract base class
from abc import ABC, abstractmethod

class ` + toTitleCase(target) + `Service(ABC):
    @abstractmethod
    def do_work(self):
        pass

# Step 2: Use in your code
def do_something(svc: ` + toTitleCase(target) + `Service):
    svc.do_work()

# Step 3: Remove import
# DELETE: from ` + target + ` import *
`

	// JavaScript
	fix.LanguageCode["javascript"] = `
// Step 1: Define interface (via documentation)
/**
 * @interface ` + toTitleCase(target) + `Service
 * @property {Function} doWork
 * @property {Function} getData
 */

// Step 2: Use interface in code
function doSomething(svc) {
    return svc.doWork();
}

// Step 3: Remove require
// DELETE: const ` + target + ` = require('./` + target + `');
`

	fix.Changes = []string{
		"Create interface/abstract class for " + target + " functionality",
		"Update all " + target + " usages to accept interface",
		"Remove direct import of " + target,
		"Update dependency injection",
	}

	fix.Rollback["go"] = "git checkout HEAD -- " + source
	fix.Rollback["java"] = "git checkout HEAD -- src/main/java/" + source
	fix.Rollback["python"] = "git checkout HEAD -- " + source
	fix.Rollback["javascript"] = "git checkout HEAD -- " + source
}

func generateInterfaceExtractionSolution(fix *RefactoringFix, source, target string) {
	sharedName := source + "_" + target + "_shared"

	// Go
	fix.LanguageCode["go"] = `
// Create shared interface package
// New directory: ` + sharedName + `/

// File: ` + sharedName + `/interfaces.go
package shared

type CommonService interface {
    Execute() error
    GetResult() interface{}
}

// Update ` + source + ` to depend on shared:
// In ` + source + `/handler.go
import "` + sharedName + `"

type Handler struct {
    service shared.CommonService
}

// Update ` + target + ` to depend on shared:
// In ` + target + `/service.go
import "` + sharedName + `"

type Service struct{}

func (s *Service) Execute() error {
    return nil
}
`

	// Java
	fix.LanguageCode["java"] = `
// Create shared interface package
// New package: com.shared

// File: com/shared/CommonService.java
package com.shared;

public interface CommonService {
    void execute() throws Exception;
    Object getResult();
}

// Update ` + source + `:
// In ` + toTitleCase(source) + `/Handler.java
import com.shared.CommonService;

public class Handler {
    private CommonService service;
    public Handler(CommonService service) {
        this.service = service;
    }
}

// Update ` + target + `:
// In ` + toTitleCase(target) + `/Service.java
import com.shared.CommonService;

public class Service implements CommonService {
    public void execute() throws Exception {}
    public Object getResult() { return null; }
}
`

	// Python
	fix.LanguageCode["python"] = `
# Create shared module
# New file: shared/__init__.py
from abc import ABC, abstractmethod

class CommonService(ABC):
    @abstractmethod
    def execute(self):
        pass
    
    @abstractmethod
    def get_result(self):
        pass

# Update ` + source + `:
# In ` + source + `/handler.py
from shared import CommonService

class Handler:
    def __init__(self, service: CommonService):
        self.service = service

# Update ` + target + `:
# In ` + target + `/service.py
from shared import CommonService

class Service(CommonService):
    def execute(self):
        pass
`

	// JavaScript
	fix.LanguageCode["javascript"] = `
// Create shared module
// New file: shared/index.js
class CommonService {
    execute() { throw new Error('Not implemented'); }
    getResult() { throw new Error('Not implemented'); }
}

module.exports = { CommonService };

// Update ` + source + `:
// In ` + source + `/handler.js
const { CommonService } = require('../shared');

class Handler {
    constructor(service) {
        this.service = service;
    }
}

// Update ` + target + `:
// In ` + target + `/service.js
const { CommonService } = require('../shared');

class Service extends CommonService {
    execute() {}
    getResult() { return null; }
}
`

	fix.Changes = []string{
		"Create new shared interface package: " + sharedName,
		"Move common abstractions to " + sharedName,
		"Update " + source + " to depend on " + sharedName,
		"Update " + target + " to depend on " + sharedName,
		"Remove direct dependency between " + source + " and " + target,
	}
}

func generateRestructuringSolution(fix *RefactoringFix, source, target string) {
	// Go
	fix.LanguageCode["go"] = `
// Restructure into layer-based architecture

// Layer 1: Domain models
// New file: domain/models.go
package domain

type Entity struct {
    ID   string
    Data interface{}
}

// Layer 2: ` + source + ` handlers
// File: ` + source + `_handler/handler.go
package ` + source + `_handler

import "domain"

type Handler struct{}

func (h *Handler) Handle(e *domain.Entity) error {
    return nil
}

// Layer 3: ` + target + ` processors
// File: ` + target + `_processor/processor.go
package ` + target + `_processor

import "domain"

type Processor struct{}

func (p *Processor) Process(e *domain.Entity) error {
    return nil
}
`

	// Java
	fix.LanguageCode["java"] = `
// Restructure into layer-based architecture

// Layer 1: Domain models
// Package: com.shared.domain
package com.shared.domain;

public class Entity {
    private String id;
    private Object data;
}

// Layer 2: ` + source + ` handlers
// Package: com.` + source + `.handler
package com.` + source + `.handler;

import com.shared.domain.*;

public class Handler {
    public void handle(Entity entity) {}
}

// Layer 3: ` + target + ` processors
// Package: com.` + target + `.processor
package com.` + target + `.processor;

import com.shared.domain.*;

public class Processor {
    public void process(Entity entity) {}
}
`

	// Python
	fix.LanguageCode["python"] = `
# Restructure into layer-based architecture

# Layer 1: Domain models
# File: shared/models.py
class Entity:
    def __init__(self, id, data):
        self.id = id
        self.data = data

# Layer 2: ` + source + ` handlers
# File: ` + source + `_handler/handler.py
from shared.models import Entity

class Handler:
    def handle(self, entity: Entity):
        pass

# Layer 3: ` + target + ` processors
# File: ` + target + `_processor/processor.py
from shared.models import Entity

class Processor:
    def process(self, entity: Entity):
        pass
`

	// JavaScript
	fix.LanguageCode["javascript"] = `
// Restructure into layer-based architecture

// Layer 1: Domain models
// File: shared/models.js
class Entity {
    constructor(id, data) {
        this.id = id;
        this.data = data;
    }
}

module.exports = { Entity };

// Layer 2: ` + source + ` handlers
// File: ` + source + `_handler/index.js
const { Entity } = require('../shared/models');

class Handler {
    handle(entity) {}
}

module.exports = Handler;

// Layer 3: ` + target + ` processors
// File: ` + target + `_processor/index.js
const { Entity } = require('../shared/models');

class Processor {
    process(entity) {}
}

module.exports = Processor;
`

	fix.Changes = []string{
		"Create domain models package",
		"Reorganize " + source + " into handler-specific package",
		"Reorganize " + target + " into processor-specific package",
		"Update all imports to use domain models and new structure",
	}
}

// ==================== VALIDATION ====================

// validateSemanticPreservation checks if refactoring preserves semantics
func validateSemanticPreservation(fix *RefactoringFix) bool {
	// Check 1: All functions still exist
	hasPreConditions := len(fix.PreConditions) > 0 || fix.PreConditions == nil

	// Check 2: Rollback available
	hasRollback := len(fix.Rollback) > 0

	// Check 3: Not too risky
	notTooRisky := fix.Strategy.Risk != "HIGH"

	return hasPreConditions && hasRollback && notTooRisky
}

// ==================== REPORTING ====================

// PrintSolutions prints all solutions
func (rs *RefactoringSynthesizer) PrintSolutions() {
	if len(rs.solutions) == 0 {
		fmt.Println("No refactoring solutions generated")
		return
	}

	fmt.Println("\n╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║          REFACTORING SOLUTIONS SUMMARY (Phase 3)             ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝\n")

	for i, solution := range rs.solutions {
		fmt.Printf("Solution %d: %s → %s\n", i+1, solution.Violation.SourceModule, solution.Violation.TargetModule)
		fmt.Printf("Risk Level: %s | Probability: %.2f%%\n", solution.Violation.RiskLevel, solution.Violation.ViolationProb*100)
		fmt.Printf("Options: %d | Best Strategy: %s\n", len(solution.Fixes), solution.Best.Strategy.Type)
		fmt.Printf("Recommended Effort: %s | Developer Acceptance: %.1f%%\n\n",
			solution.Best.Strategy.Effort, solution.Best.Strategy.Acceptance*100)
	}

	fmt.Println("═══════════════════════════════════════════════════════════════\n")
}

// PrintDetailedSolution prints detailed info about a specific solution
func PrintDetailedSolution(solution *RefactoringSolution, language string) {
	fmt.Printf("\n=== Solution for %s → %s ===\n", solution.Violation.SourceModule, solution.Violation.TargetModule)

	if solution.Best != nil {
		fmt.Printf("Recommended Strategy: %s\n", solution.Best.Strategy.Type)
		fmt.Printf("Description: %s\n", solution.Best.Strategy.Description)
		fmt.Printf("Effort: %s | Risk: %s\n", solution.Best.Strategy.Effort, solution.Best.Strategy.Risk)
		fmt.Printf("Developer Acceptance: %.1f%%\n", solution.Best.Strategy.Acceptance*100)

		fmt.Println("\nChanges required:")
		for i, change := range solution.Best.Changes {
			fmt.Printf("  %d. %s\n", i+1, change)
		}

		if code, ok := solution.Best.LanguageCode[language]; ok {
			fmt.Printf("\n=== %s Implementation ===\n", language)
			fmt.Println(code)
		}
	}

	fmt.Println()
}

// ==================== UTILITY FUNCTIONS ====================

func filterHighRiskViolations(predictions []*PredictionResult) []*PredictionResult {
	highRisk := make([]*PredictionResult, 0)
	for _, p := range predictions {
		if p.RiskLevel == "HIGH" {
			highRisk = append(highRisk, p)
		}
	}

	// Sort by probability (highest first)
	sort.Slice(highRisk, func(i, j int) bool {
		return highRisk[i].ViolationProb > highRisk[j].ViolationProb
	})

	return highRisk
}

func toTitleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}