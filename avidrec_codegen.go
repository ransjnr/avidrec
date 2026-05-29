package main

import (
	"fmt"
	"strings"
)

// RefactoringStrategy represents a potential fix for a violation
type RefactoringStrategy struct {
	Type        string  // DEPENDENCY_REMOVAL, INTERFACE_EXTRACTION, MODULE_RESTRUCTURING
	Description string  // Human-readable description
	Cost        float64 // Estimated refactoring cost (0-1000)
	Effort      string  // LOW, MEDIUM, HIGH
	Risk        string  // LOW, MEDIUM, HIGH
	Acceptance  float64 // Estimated developer acceptance (0-1.0)
}

// RefactoringFix represents the actual refactoring code
type RefactoringFix struct {
	Strategy       *RefactoringStrategy
	SourceModule   string
	TargetModule   string
	LanguageCode   map[string]string // Language -> Code snippets (go, java, python, javascript)
	Changes        []string          // List of changes to make
	PreConditions  []string          // Must be true before applying fix
	PostConditions []string          // Will be true after applying fix
	Rollback       map[string]string  // Rollback code per language
	Score          float64           // Quality score
}

// CodeGenerator generates multi-language refactoring code
type CodeGenerator struct {
	violations []*PredictionResult
	graph      *DependencyGraph
	solver     *Z3Solver
}

// NewCodeGenerator creates a new code generator
func NewCodeGenerator(violations []*PredictionResult, graph *DependencyGraph) *CodeGenerator {
	return &CodeGenerator{
		violations: violations,
		graph:      graph,
		solver:     NewZ3Solver(),
	}
}

// GenerateRefactoringFixes generates fixes for all high-risk violations
func (cg *CodeGenerator) GenerateRefactoringFixes() []*RefactoringFix {
	fixes := make([]*RefactoringFix, 0)

	for _, violation := range cg.violations {
		if violation.RiskLevel != "HIGH" {
			continue // Only fix high-risk violations
		}

		// Extract features for constraint solving
		extractor := NewFeatureExtractor(cg.graph)
		features := extractor.ExtractFeaturesForDependency(violation.SourceModule, violation.TargetModule)

		// Get refactoring strategies from Z3
		strategies := cg.solver.SolveRefactoringOptions(violation, features)

		// Generate code for each strategy
		for _, strategy := range strategies {
			fix := cg.generateFix(violation, strategy, features)
			if fix != nil && fix.Score > 0.5 {
				fixes = append(fixes, fix)
			}
		}
	}

	return fixes
}

// generateFix generates refactoring code for a single violation+strategy
func (cg *CodeGenerator) generateFix(violation *PredictionResult, strategy *RefactoringStrategy, features *DependencyFeatures) *RefactoringFix {
	fix := &RefactoringFix{
		Strategy:      strategy,
		SourceModule:  violation.SourceModule,
		TargetModule:  violation.TargetModule,
		LanguageCode:  make(map[string]string),
		Rollback:      make(map[string]string),
		Changes:       make([]string, 0),
		Score:         strategy.Acceptance,
	}

	switch strategy.Type {
	case "DEPENDENCY_REMOVAL":
		cg.generateRemovalFix(fix)
	case "INTERFACE_EXTRACTION":
		cg.generateInterfaceExtractionFix(fix)
	case "MODULE_RESTRUCTURING":
		cg.generateRestructuringFix(fix)
	}

	return fix
}

// ==================== DEPENDENCY REMOVAL ====================

func (cg *CodeGenerator) generateRemovalFix(fix *RefactoringFix) {
	source := extractPackageName(fix.SourceModule)
	target := extractPackageName(fix.TargetModule)

	// Go code
	fix.LanguageCode["go"] = fmt.Sprintf(`
// REFACTORING: Remove dependency %s → %s

// Before:
// import "%s"
//
// func SomeFunc() {
//     %s.DoSomething()
// }

// After:
// Remove the import statement above
// Replace %s.DoSomething() calls with local implementation or interface call

// Step 1: Find all usages of %s in %s
// grep -r "%s\." %s/

// Step 2: Replace with interface or local implementation
// See interface_extraction.go for interface-based approach

// Step 3: Remove import
// Delete: import "%s"
`, fix.SourceModule, fix.TargetModule, fix.TargetModule, target, target, target, source, target, source, fix.TargetModule)

	fix.LanguageCode["go"] += cg.generateGoRemovalCode(source, target)

	// Java code
	fix.LanguageCode["java"] = cg.generateJavaRemovalCode(source, target)

	// Python code
	fix.LanguageCode["python"] = cg.generatePythonRemovalCode(source, target)

	// JavaScript code
	fix.LanguageCode["javascript"] = cg.generateJavaScriptRemovalCode(source, target)

	// Changes
	fix.Changes = []string{
		fmt.Sprintf("Remove import of %s from %s", fix.TargetModule, fix.SourceModule),
		fmt.Sprintf("Replace %s calls with local implementation or interface", target),
		fmt.Sprintf("Update any dependent code in %s", fix.SourceModule),
	}

	// Rollback
	fix.Rollback["go"] = fmt.Sprintf("git checkout HEAD -- %s/", source)
}

func (cg *CodeGenerator) generateGoRemovalCode(source, target string) string {
	return fmt.Sprintf(`
// Go implementation for removing dependency
package %s

// Step-by-step removal:

// 1. Create an interface for the functionality you were using from %s
type %sInterface interface {
    DoSomething() error
    GetData() interface{}
}

// 2. Update your functions to accept interface instead of concrete type
// Before:
// func MyFunc(%s *%s.Service) error {
//     return %s.DoSomething()
// }

// After:
func MyFunc(service %sInterface) error {
    return service.DoSomething()
}

// 3. At injection points, pass the interface implementation
// Remove: import "%s"
// Keep injecting the interface implementation
`, source, target, strings.Title(target), target, strings.Title(target), target, strings.Title(target), target)
}

func (cg *CodeGenerator) generateJavaRemovalCode(source, target string) string {
	return fmt.Sprintf(`
// Java implementation for removing dependency
package %s;

// Step 1: Create an interface for the dependency
public interface %sService {
    void doSomething() throws Exception;
    Object getData();
}

// Step 2: Update your class to use the interface
// Before:
// import %s.%sImpl;
// public class %s {
//     private %sImpl service;
// }

// After:
public class %s {
    private %sService service;
    
    public %s(%sService service) {
        this.service = service;
    }
}

// Step 3: Remove the import
// Delete: import %s.%sImpl;
`, source, strings.Title(target), target, strings.Title(target), strings.Title(source), strings.Title(target), strings.Title(source), strings.Title(target), strings.Title(source), strings.Title(target), target, strings.Title(target))
}

func (cg *CodeGenerator) generatePythonRemovalCode(source, target string) string {
	return fmt.Sprintf(`
# Python implementation for removing dependency
# File: %s/__init__.py

# Step 1: Create an abstract base class
from abc import ABC, abstractmethod

class %sInterface(ABC):
    @abstractmethod
    def do_something(self):
        pass
    
    @abstractmethod
    def get_data(self):
        pass

# Step 2: Update your class to use the interface
# Before:
# from %s import %s
# class MyClass:
#     def __init__(self):
#         self.service = %s.Service()

# After:
class MyClass:
    def __init__(self, service: %sInterface):
        self.service = service
    
    def my_method(self):
        self.service.do_something()

# Step 3: Remove the import
# Delete: from %s import %s
`, source, strings.Title(target), target, strings.Title(target), strings.Title(target), strings.Title(target), target, strings.Title(target))
}

func (cg *CodeGenerator) generateJavaScriptRemovalCode(source, target string) string {
	return fmt.Sprintf(`
// JavaScript implementation for removing dependency
// File: %s/index.js

// Step 1: Create an interface (using documentation)
/**
 * @interface %sService
 * @property {Function} doSomething - Do something
 * @property {Function} getData - Get data
 */

// Step 2: Update your class to use the interface
// Before:
// const %s = require('./%s');
// class MyClass {
//     constructor() {
//         this.service = new %s.Service();
//     }
// }

// After:
class MyClass {
    /**
     * @param {%sService} service
     */
    constructor(service) {
        this.service = service;
    }
    
    myMethod() {
        return this.service.doSomething();
    }
}

// Step 3: Remove the require statement
// Delete: const %s = require('./%s');

module.exports = MyClass;
`, source, strings.Title(target), strings.Title(target), target, strings.Title(target), strings.Title(target), target, target)
}

// ==================== INTERFACE EXTRACTION ====================

func (cg *CodeGenerator) generateInterfaceExtractionFix(fix *RefactoringFix) {
	source := extractPackageName(fix.SourceModule)
	target := extractPackageName(fix.TargetModule)

	// Go code
	fix.LanguageCode["go"] = cg.generateGoInterfaceCode(source, target)

	// Java code
	fix.LanguageCode["java"] = cg.generateJavaInterfaceCode(source, target)

	// Python code
	fix.LanguageCode["python"] = cg.generatePythonInterfaceCode(source, target)

	// JavaScript code
	fix.LanguageCode["javascript"] = cg.generateJavaScriptInterfaceCode(source, target)

	fix.Changes = []string{
		fmt.Sprintf("Create new interface package for shared functionality between %s and %s", source, target),
		fmt.Sprintf("Move shared abstractions to interface package"),
		fmt.Sprintf("Update both %s and %s to depend on interface instead of each other", source, target),
	}
}

func (cg *CodeGenerator) generateGoInterfaceCode(source, target string) string {
	return fmt.Sprintf(`
// Go: Create interface layer to break circular dependency

// New file: shared/interfaces.go
package shared

type ServiceInterface interface {
    Execute() error
    GetResult() interface{}
}

// Update %s to depend on shared:
// File: %s/handler.go
package %s

import "shared"

type Handler struct {
    service shared.ServiceInterface
}

func (h *Handler) Process() error {
    return h.service.Execute()
}

// Update %s to implement ServiceInterface:
// File: %s/service.go
package %s

import "shared"

type Service struct{}

func (s *Service) Execute() error {
    // implementation
    return nil
}

func (s *Service) GetResult() interface{} {
    return nil
}

// Now: %s → shared ← %s (no circular dependency!)
`, source, source, source, target, target, target, source, target)
}

func (cg *CodeGenerator) generateJavaInterfaceCode(source, target string) string {
	return fmt.Sprintf(`
// Java: Create interface layer to break circular dependency

// New file: shared/ServiceInterface.java
package shared;

public interface ServiceInterface {
    void execute() throws Exception;
    Object getResult();
}

// Update %s:
// File: %s/Handler.java
package %s;

import shared.ServiceInterface;

public class Handler {
    private ServiceInterface service;
    
    public Handler(ServiceInterface service) {
        this.service = service;
    }
    
    public void process() throws Exception {
        service.execute();
    }
}

// Update %s to implement ServiceInterface:
// File: %s/Service.java
package %s;

import shared.ServiceInterface;

public class Service implements ServiceInterface {
    public void execute() throws Exception {
        // implementation
    }
    
    public Object getResult() {
        return null;
    }
}

// Dependency structure:
// %s → shared ← %s
`, source, source, source, target, target, target, source, target)
}

func (cg *CodeGenerator) generatePythonInterfaceCode(source, target string) string {
	return fmt.Sprintf(`
# Python: Create interface layer to break circular dependency

# New file: shared/__init__.py
from abc import ABC, abstractmethod

class ServiceInterface(ABC):
    @abstractmethod
    def execute(self):
        pass
    
    @abstractmethod
    def get_result(self):
        pass

# Update %s/handler.py
from shared import ServiceInterface

class Handler:
    def __init__(self, service: ServiceInterface):
        self.service = service
    
    def process(self):
        self.service.execute()

# Update %s/service.py
from shared import ServiceInterface

class Service(ServiceInterface):
    def execute(self):
        pass
    
    def get_result(self):
        return None

# Dependency structure:
# %s → shared ← %s
`, source, target, source, target)
}

func (cg *CodeGenerator) generateJavaScriptInterfaceCode(source, target string) string {
	return fmt.Sprintf(`
// JavaScript: Create interface layer to break circular dependency

// New file: shared/ServiceInterface.js
/**
 * @interface ServiceInterface
 * @property {Function} execute
 * @property {Function} getResult
 */

class ServiceInterface {
    execute() {
        throw new Error('Not implemented');
    }
    
    getResult() {
        throw new Error('Not implemented');
    }
}

module.exports = ServiceInterface;

// Update %s/handler.js
const ServiceInterface = require('../shared/ServiceInterface');

class Handler {
    constructor(service) {
        if (!(service instanceof ServiceInterface)) {
            throw new Error('service must implement ServiceInterface');
        }
        this.service = service;
    }
    
    process() {
        return this.service.execute();
    }
}

module.exports = Handler;

// Update %s/service.js
const ServiceInterface = require('../shared/ServiceInterface');

class Service extends ServiceInterface {
    execute() {
        // implementation
    }
    
    getResult() {
        return null;
    }
}

module.exports = Service;

// Dependency structure:
// %s → shared ← %s
`, source, target, source, target)
}

// ==================== MODULE RESTRUCTURING ====================

func (cg *CodeGenerator) generateRestructuringFix(fix *RefactoringFix) {
	source := extractPackageName(fix.SourceModule)
	target := extractPackageName(fix.TargetModule)

	fix.LanguageCode["go"] = cg.generateGoRestructuringCode(source, target)
	fix.LanguageCode["java"] = cg.generateJavaRestructuringCode(source, target)
	fix.LanguageCode["python"] = cg.generatePythonRestructuringCode(source, target)
	fix.LanguageCode["javascript"] = cg.generateJavaScriptRestructuringCode(source, target)

	fix.Changes = []string{
		fmt.Sprintf("Analyze module structure of %s and %s", source, target),
		fmt.Sprintf("Identify overlapping responsibilities"),
		fmt.Sprintf("Reorganize into cleaner dependency structure"),
	}
}

func (cg *CodeGenerator) generateGoRestructuringCode(source, target string) string {
	return fmt.Sprintf(`
// Go: Module Restructuring Strategy

// Current (Problematic) Structure:
// %s/ → imports → %s/
// %s/ → imports → %s/

// Proposed Restructuring:

// 1. Create core package with shared logic
// New file: core/domain.go
package core

type Entity struct {
    ID    string
    Name  string
}

// 2. Move %s responsibility
// File: %s_handler/handler.go
package %s_handler

import "core"

type Handler struct{}

func (h *Handler) Handle(e *core.Entity) error {
    return nil
}

// 3. Move %s responsibility
// File: %s_processor/processor.go
package %s_processor

import "core"

type Processor struct{}

func (p *Processor) Process(e *core.Entity) error {
    return nil
}

// New Structure:
// %s_handler → core ← %s_processor
`, source, target, target, source, source, source, source, target, target, target, source, target)
}

func (cg *CodeGenerator) generateJavaRestructuringCode(source, target string) string {
	return fmt.Sprintf(`
// Java: Module Restructuring Strategy

// Current structure:
// com.%s.* → imports → com.%s.*
// com.%s.* → imports → com.%s.*

// Proposed: Move to layer-based architecture

// Layer 1: Core Domain
package com.shared.domain;

public class Entity {
    private String id;
    private String name;
}

// Layer 2: %s Handlers
package com.%s.handler;

import com.shared.domain.*;

public class Handler {
    public void handle(Entity entity) {
        // implementation
    }
}

// Layer 3: %s Processors
package com.%s.processor;

import com.shared.domain.*;

public class Processor {
    public void process(Entity entity) {
        // implementation
    }
}

// New dependency structure (no cycles):
// com.%s.* → com.shared.* ← com.%s.*
`, source, target, target, source, source, source, target, target, source, target)
}

func (cg *CodeGenerator) generatePythonRestructuringCode(source, target string) string {
	return fmt.Sprintf(`
# Python: Module Restructuring Strategy

# Current (problematic):
# from %s import ...
# from %s import ...
# (creating cycle)

# Proposed: Layer-based architecture

# Layer 1: Shared domain
# File: shared/models.py
class Entity:
    def __init__(self, id, name):
        self.id = id
        self.name = name

# Layer 2: %s functionality
# File: %s/handler.py
from shared.models import Entity

class Handler:
    def handle(self, entity: Entity):
        pass

# Layer 3: %s functionality
# File: %s/processor.py
from shared.models import Entity

class Processor:
    def process(self, entity: Entity):
        pass

# New structure (acyclic):
# %s → shared ← %s
`, source, target, source, source, target, target, source, target)
}

func (cg *CodeGenerator) generateJavaScriptRestructuringCode(source, target string) string {
	return fmt.Sprintf(`
// JavaScript: Module Restructuring Strategy

// Current (problematic) structure
// %s imports from %s
// %s imports from %s

// Proposed: Layer-based architecture

// Layer 1: Shared models
// File: shared/models.js
class Entity {
    constructor(id, name) {
        this.id = id;
        this.name = name;
    }
}

module.exports = { Entity };

// Layer 2: %s functionality
// File: %s/handler.js
const { Entity } = require('../shared/models');

class Handler {
    handle(entity) {
        // implementation
    }
}

module.exports = Handler;

// Layer 3: %s functionality
// File: %s/processor.js
const { Entity } = require('../shared/models');

class Processor {
    process(entity) {
        // implementation
    }
}

module.exports = Processor;

// New structure (acyclic):
// %s → shared ← %s
`, source, target, target, source, source, source, target, target, source, target)
}

// ==================== UTILITY FUNCTIONS ====================

func extractPackageName(modulePath string) string {
	parts := strings.Split(modulePath, "/")
	return parts[len(parts)-1]
}

// PrintFix prints a refactoring fix in readable format
func PrintFix(fix *RefactoringFix) {
	fmt.Printf("\n=== Refactoring Fix ===\n")
	fmt.Printf("Strategy: %s\n", fix.Strategy.Type)
	fmt.Printf("Description: %s\n", fix.Strategy.Description)
	fmt.Printf("Effort: %s | Risk: %s | Acceptance: %.1f%%\n",
		fix.Strategy.Effort, fix.Strategy.Risk, fix.Strategy.Acceptance*100)
	fmt.Printf("Score: %.2f\n", fix.Score)

	fmt.Println("\nChanges:")
	for i, change := range fix.Changes {
		fmt.Printf("  %d. %s\n", i+1, change)
	}

	fmt.Println("\nSupported Languages:")
	for lang := range fix.LanguageCode {
		fmt.Printf("  ✓ %s\n", strings.ToUpper(lang))
	}

	fmt.Println()
}

// PrintFixCode prints the generated code for a specific language
func PrintFixCode(fix *RefactoringFix, language string) {
	if code, ok := fix.LanguageCode[language]; ok {
		fmt.Printf("\n=== %s Code ===\n", strings.ToUpper(language))
		fmt.Println(code)
	} else {
		fmt.Printf("No %s code available for this fix\n", language)
	}
}