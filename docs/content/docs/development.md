---
title: "Development Guide"
linkTitle: "Development"
weight: 5
description: >
  Advanced development patterns and architecture in sync-tools v0.4.0
---

# Development Guide - v0.4.0 Architecture

This guide covers the advanced development patterns introduced in sync-tools v0.4.0, including the Type State Pattern, enhanced Test Driver, and Object Mother patterns.

## Architecture Overview

sync-tools v0.4.0 introduces several sophisticated patterns that provide compile-time safety, enhanced testability, and domain-driven design:

- **Type State Pattern**: Compile-time guarantees for state transitions
- **Enhanced Test Driver**: Fluent API for complex test scenarios  
- **Object Mother Patterns**: Domain-specific builders for different user personas
- **Multi-Persona Design**: Tailored interfaces for DevOps, Developers, and Compliance teams

## Type State Pattern

The Type State Pattern ensures that state transitions in the interactive wizard are validated at compile-time, preventing invalid state transitions and providing better developer experience.

### Implementation

```go
// Define valid transition types as concrete structs
type InitialToSourceSelection struct{ 
    StateTransition[InitialState, SourceSelectionState] 
}

type SourceToDestination struct{ 
    StateTransition[SourceSelectionState, DestinationSelectionState] 
}

// Type-safe state transition helpers
func (sm *StateMachine) FromInitialToSourceSelection(newState SourceSelectionState) error {
    // Compile-time guarantee: this method can only be called with SourceSelectionState
    return sm.TransitionTo(newState)
}

func (sm *StateMachine) FromSourceToDestination(newState DestinationSelectionState) error {
    // Compile-time guarantee: this method can only be called with DestinationSelectionState
    return sm.TransitionTo(newState)
}
```

### Benefits

- **Compile-time Safety**: Invalid state transitions are caught at build time
- **Self-documenting Code**: Valid transitions are explicitly defined
- **IDE Support**: Better autocomplete and type checking
- **Maintainability**: Changes to state machine structure are validated

### Usage Example

```go
// ✅ Valid transition - compiles
stateMachine.FromInitialToSourceSelection(sourceState)

// ❌ Invalid transition - compilation error
// stateMachine.FromInitialToSourceSelection(destinationState) // Won't compile!
```

## Enhanced Test Driver

The Test Driver provides a fluent API for building complex test scenarios with built-in expectations and validation.

### Fluent API Design

```go
// Fluent scenario building
result := driver.NewScenario().
    WithSource("./test_source").
    WithDestination("./test_dest").
    WithMode("two-way").
    WithExclusionPatterns("*.tmp", "*.log").
    WithGitIgnoreEnabled().
    ExpectSuccess().
    ExpectSyncFileContaining("MODE two-way").
    ExecuteInTestMode()

// Validation happens automatically
if !result.Success {
    t.Errorf("Scenario failed: %s", result.Error)
}
```

### Type-safe State Management

```go
type WizardScenarioBuilder interface {
    // Configuration methods return the builder for chaining
    WithSource(path string) WizardScenarioBuilder
    WithDestination(path string) WizardScenarioBuilder
    WithMode(mode string) WizardScenarioBuilder
    
    // Expectation methods for test validation
    ExpectSuccess() WizardScenarioBuilder
    ExpectFailure(expectedError string) WizardScenarioBuilder
    ExpectSyncFileContaining(content string) WizardScenarioBuilder
    
    // Execution methods
    Execute() *WizardResult
    ExecuteInTestMode() *WizardResult
}
```

### Advanced Scenarios

```go
// Complex multi-step scenario
driver.NewScenario().
    WithSourceFiles(map[string]string{
        "config.yml": "version: 1.0",
        "app.log":    "application logs",
    }).
    WithDestinationFiles(map[string]string{
        "config.yml": "version: 0.9",
    }).
    WithMode("two-way").
    WithExclusionPatterns("*.log").
    ExpectSyncFileContaining("EXCLUDE *.log").
    ExpectSuccess().
    ExecuteInTestMode()
```

## Object Mother Patterns

Object Mother patterns provide pre-built configurations for different user personas and use cases, embodying domain knowledge in reusable builders.

### Persona-Specific Builders

#### DevOps Scenarios

```go
// DevOps deployment scenario
config := DevOps.DeploymentSync("./app", "./prod")
// Pre-configured with:
// - One-way sync mode
// - Exclusions: *.tmp, *.log, .git/, node_modules/, target/
// - GitIgnore enabled

// DevOps backup scenario  
config := DevOps.BackupSync("./data", "./backup")
// Pre-configured with:
// - Two-way sync mode
// - Exclusions: *.cache, *.lock
// - GitIgnore disabled (include all files)
```

#### Developer Scenarios

```go
// Developer project sync
config := Developer.ProjectSync("./src", "./staging")
// Pre-configured with:
// - One-way sync mode  
// - Exclusions: node_modules/, target/, build/, dist/, *.o, *.class
// - GitIgnore enabled

// Asset synchronization
config := Developer.AssetSync("./assets", "./cdn")
// Pre-configured with:
// - Two-way sync mode
// - Minimal exclusions: *.tmp
```

#### Compliance Scenarios

```go
// Compliance audit sync
config := Compliance.AuditSync("./records", "./archive") 
// Pre-configured with:
// - One-way sync mode
// - Dry-run enabled (compliance requires review)
// - GitIgnore disabled (include all files for compliance)
```

### Advanced Product Manager Patterns

```go
// Product Manager specific workflows
pmWizard := ProductManagerWizard()

// Feature deployment with review
config := pmWizard.FeatureDeployment("./feature", "./staging")
// Pre-configured with:
// - Exclusions: *.test.*, *.spec.*, test/, spec/
// - Dry-run enabled (PM always wants to review first)

// Emergency hotfix (faster deployment)
config := pmWizard.HotfixDeployment("./hotfix", "./production")  
// Pre-configured with:
// - Dry-run disabled (hotfixes need to be fast)
// - Minimal exclusions
```

### Multi-Stage Scenarios

```go
// Complex multi-stage deployment
stages := NewMultiStageWizard().
    Stage1().
        WithSourceDir("./src").
        WithDestinationDir("./build").
        WithMode("one-way").
        Build().
    Stage2(stage1Result).
        WithSourceDir("./build").  
        WithDestinationDir("./staging").
        WithMode("one-way").
        Build().
    Stage3(stage1Result, stage2Result).
        WithSourceDir("./staging").
        WithDestinationDir("./production").
        WithMode("one-way").
        WithDryRun(true). // Final stage requires approval
        Build().
    Build()

// Execute all stages
for i, stage := range stages {
    result := wizard.ExecuteStage(stage)
    if !result.Success {
        log.Fatalf("Stage %d failed: %s", i+1, result.Error)
    }
}
```

## BDD/TDD Integration

### Red/Green/Refactor Workflow

The architecture supports strict BDD/TDD discipline:

```go
// 1. RED: Write failing BDD scenario first
func TestWizardGeneratesTwoWaySyncFile(t *testing.T) {
    // Given a wizard scenario
    scenario := driver.NewScenario().
        WithSource("./test_src").
        WithDestination("./test_dst").  
        WithMode("two-way").
        ExpectSyncFileContaining("MODE two-way")
    
    // When executed in test mode
    result := scenario.ExecuteInTestMode()
    
    // Then it should succeed
    assert.True(t, result.Success)
}

// 2. GREEN: Implement minimal code to make test pass
// (Implementation in wizard package)

// 3. REFACTOR: Clean up code while maintaining green tests
// (Refactor using type-safe patterns and object mothers)
```

### Living Documentation

Object Mothers serve as executable specifications:

```go
// These patterns document actual user workflows
func TestDevOpsDeploymentWorkflow(t *testing.T) {
    // This test documents how DevOps teams actually use sync-tools
    config := DevOps.DeploymentSync("./app", "./production")
    
    driver := testdriver.NewWizardDriver()
    result := driver.GenerateSyncFile(config)
    
    // Expectations based on real DevOps requirements
    assert.Contains(t, result.SyncFileContent, "EXCLUDE node_modules/")
    assert.Contains(t, result.SyncFileContent, "GITIGNORE true")
    assert.Contains(t, result.SyncFileContent, "MODE one-way")
}
```

## Testing Best Practices

### Type-Safe Test Construction

```go
// Use type-safe builders to prevent test configuration errors
func TestComplexSyncScenario(t *testing.T) {
    // Builder prevents invalid combinations at compile time
    scenario := mother.WizardConfigFor("./src", "./dst").
        WithMode("two-way").
        WithExclusionPatterns("*.tmp", "*.log"). 
        WithGitIgnore(true).
        WithDryRun(false).
        Build()
    
    // Type-safe execution
    result := driver.GenerateSyncFile(scenario)
    
    // Fluent assertions
    assert.True(t, result.Success)
    assert.Contains(t, result.SyncFileContent, "MODE two-way")
}
```

### Persona-Driven Test Coverage

```go
// Test coverage organized by user personas
func TestDevOpsPersonaWorkflows(t *testing.T) {
    testCases := []struct {
        name   string
        config *wizard.TestModeOptions
    }{
        {"deployment", DevOps.DeploymentSync("./app", "./prod")},
        {"backup", DevOps.BackupSync("./data", "./backup")},
        {"configuration", DevOps.ConfigurationSync("./config", "./remote")},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := driver.GenerateSyncFile(tc.config)
            assert.True(t, result.Success, "DevOps scenario should succeed")
        })
    }
}
```

## Contributing to v0.4.0 Architecture

### Adding New Persona Patterns

1. **Create Persona Builder**:
```go
type DataScientistScenarios struct{}

func (d DataScientistScenarios) NotebookSync(sourceDir, destDir string) *wizard.TestModeOptions {
    return NewWizardConfig().
        WithSourceDir(sourceDir).
        WithDestinationDir(destDir).
        WithMode("two-way").
        WithExclusionPatterns("*.pyc", "__pycache__/", ".ipynb_checkpoints/").
        WithGitIgnore(true).
        Build()
}
```

2. **Add to Convenience Exports**:
```go
var (
    DevOps      = DevOpsScenarios{}
    Developer   = DeveloperScenarios{}
    Compliance  = ComplianceScenarios{}
    DataScience = DataScientistScenarios{} // New persona
)
```

3. **Write BDD Scenarios**:
```gherkin
Feature: Data Science Workflows
  As a data scientist
  I want to sync notebooks and datasets
  So that I can collaborate effectively

  Scenario: Notebook synchronization
    Given I have Jupyter notebooks in "./research"
    When I use DataScience.NotebookSync("./research", "./shared")
    Then it should exclude checkpoint directories
    And it should enable two-way synchronization
```

### Extending State Machine

1. **Define New States**:
```go
type DataValidationState struct {
    // State-specific fields
}

func (d DataValidationState) GetStateName() string {
    return "DataValidation"
}
```

2. **Add Type-Safe Transitions**:
```go
type ProgressToDataValidation struct{ 
    StateTransition[ProgressState, DataValidationState] 
}

func (sm *StateMachine) FromProgressToDataValidation(newState DataValidationState) error {
    return sm.TransitionTo(newState)
}
```

3. **Implement State Operations**:
```go
func (sm *StateMachine) GetDataValidationOperations() (*DataValidationOperations, error) {
    if state, ok := sm.currentState.(DataValidationState); ok {
        return &DataValidationOperations{sm: sm, state: state}, nil
    }
    return nil, fmt.Errorf("not in data validation state")
}
```

This architecture ensures that sync-tools v0.4.0 provides a robust, type-safe foundation for complex synchronization workflows while maintaining excellent testability and domain-driven design principles.