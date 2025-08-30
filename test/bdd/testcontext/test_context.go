package testcontext

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/DamianReeves/sync-tools/test/bdd/driver"
	"github.com/DamianReeves/sync-tools/test/bdd/mother"
)

// TestEnvironment provides a clean testing environment with proper isolation
type TestEnvironment struct {
	// Directories
	TempRoot   string
	SourceDir  string
	DestDir    string
	WorkingDir string
	
	// Driver for sync-tools interaction
	Driver driver.SyncDriver
	
	// Last command result for assertions
	LastResult *driver.SyncResult
}

// NewTestEnvironment creates a new isolated test environment
func NewTestEnvironment(binaryPath string) (*TestEnvironment, error) {
	// Create temporary directory using proper temp file tools
	tempRoot, err := os.MkdirTemp("", "sync-tools-test-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	
	env := &TestEnvironment{
		TempRoot:   tempRoot,
		SourceDir:  filepath.Join(tempRoot, "source"),
		DestDir:    filepath.Join(tempRoot, "dest"),
		WorkingDir: filepath.Join(tempRoot, "work"),
		Driver:     driver.NewSyncDriver(binaryPath),
	}
	
	env.Driver.SetWorkingDir(env.WorkingDir)
	
	// Create working directory
	if err := os.MkdirAll(env.WorkingDir, 0755); err != nil {
		env.Cleanup()
		return nil, fmt.Errorf("failed to create working directory: %w", err)
	}
	
	return env, nil
}

// Cleanup removes all temporary files and directories
func (env *TestEnvironment) Cleanup() {
	if env.TempRoot != "" {
		_ = os.RemoveAll(env.TempRoot)
	}
}

// SetupDirectories creates source and destination directories using Object Mother patterns
func (env *TestEnvironment) SetupDirectories(sourceBuilder, destBuilder mother.DirectoryBuilder) error {
	if err := sourceBuilder.Build(env.SourceDir); err != nil {
		return fmt.Errorf("failed to build source directory: %w", err)
	}
	
	if err := destBuilder.Build(env.DestDir); err != nil {
		return fmt.Errorf("failed to build dest directory: %w", err)
	}
	
	return nil
}

// Sync operations using Test Driver
func (env *TestEnvironment) Sync(options ...driver.SyncOption) *driver.SyncResult {
	env.LastResult = env.Driver.Sync(env.SourceDir, env.DestDir, options...)
	return env.LastResult
}

func (env *TestEnvironment) GeneratePlan(planFile string, options ...driver.PlanOption) *driver.SyncResult {
	planPath := filepath.Join(env.WorkingDir, planFile)
	env.LastResult = env.Driver.GeneratePlan(env.SourceDir, env.DestDir, planPath, options...)
	return env.LastResult
}

func (env *TestEnvironment) ApplyPlan(planFile string, options ...driver.ApplyOption) *driver.SyncResult {
	planPath := filepath.Join(env.WorkingDir, planFile)
	env.LastResult = env.Driver.ApplyPlan(planPath, options...)
	return env.LastResult
}

func (env *TestEnvironment) GeneratePatch(patchFile string, options ...driver.PatchOption) *driver.SyncResult {
	patchPath := filepath.Join(env.WorkingDir, patchFile)
	env.LastResult = env.Driver.GeneratePatch(env.SourceDir, env.DestDir, patchPath, options...)
	return env.LastResult
}

// File system assertions and utilities
func (env *TestEnvironment) FileExists(relativePath string) bool {
	fullPath := filepath.Join(env.WorkingDir, relativePath)
	_, err := os.Stat(fullPath)
	return err == nil
}

func (env *TestEnvironment) FileContent(relativePath string) (string, error) {
	fullPath := filepath.Join(env.WorkingDir, relativePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (env *TestEnvironment) SourceFileExists(relativePath string) bool {
	fullPath := filepath.Join(env.SourceDir, relativePath)
	_, err := os.Stat(fullPath)
	return err == nil
}

func (env *TestEnvironment) DestFileExists(relativePath string) bool {
	fullPath := filepath.Join(env.DestDir, relativePath)
	_, err := os.Stat(fullPath)
	return err == nil
}

func (env *TestEnvironment) SourceFileContent(relativePath string) (string, error) {
	fullPath := filepath.Join(env.SourceDir, relativePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (env *TestEnvironment) DestFileContent(relativePath string) (string, error) {
	fullPath := filepath.Join(env.DestDir, relativePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// Assertion helpers
func (env *TestEnvironment) AssertLastCommandSucceeded() error {
	if env.LastResult == nil {
		return fmt.Errorf("no command has been executed")
	}
	if !env.LastResult.Success {
		return fmt.Errorf("expected command to succeed (exit code 0), but got %d. Output: %s", 
			env.LastResult.ExitCode, env.LastResult.Output)
	}
	return nil
}

func (env *TestEnvironment) AssertLastCommandFailed() error {
	if env.LastResult == nil {
		return fmt.Errorf("no command has been executed")
	}
	if env.LastResult.Success {
		return fmt.Errorf("expected command to fail, but it succeeded. Output: %s", env.LastResult.Output)
	}
	return nil
}

func (env *TestEnvironment) AssertFileContains(relativePath, expectedContent string) error {
	content, err := env.FileContent(relativePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", relativePath, err)
	}
	
	if content != expectedContent {
		return fmt.Errorf("file %s content mismatch.\nExpected: %s\nActual: %s", 
			relativePath, expectedContent, content)
	}
	
	return nil
}

func (env *TestEnvironment) AssertOutputContains(expectedText string) error {
	if env.LastResult == nil {
		return fmt.Errorf("no command has been executed")
	}
	
	if !contains(env.LastResult.Output, expectedText) {
		return fmt.Errorf("expected output to contain '%s', but got: %s", 
			expectedText, env.LastResult.Output)
	}
	
	return nil
}

func (env *TestEnvironment) AssertOutputDoesNotContain(unexpectedText string) error {
	if env.LastResult == nil {
		return fmt.Errorf("no command has been executed")
	}
	
	if contains(env.LastResult.Output, unexpectedText) {
		return fmt.Errorf("expected output to not contain '%s', but it does. Output: %s", 
			unexpectedText, env.LastResult.Output)
	}
	
	return nil
}

// Helper function for substring checking
func contains(s, substr string) bool {
	return len(substr) == 0 || len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		containsAt(s, substr))))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ExecuteRawCommand executes a raw command string with path replacement
func (env *TestEnvironment) ExecuteRawCommand(args string) error {
	// Replace test path placeholders with actual directories
	updatedArgs := args
	updatedArgs = strings.ReplaceAll(updatedArgs, "./test_source", env.SourceDir)
	updatedArgs = strings.ReplaceAll(updatedArgs, "./test_dest", env.DestDir)
	updatedArgs = strings.ReplaceAll(updatedArgs, "./source", env.SourceDir)
	updatedArgs = strings.ReplaceAll(updatedArgs, "./dest", env.DestDir)
	
	// Execute command through driver
	err := env.Driver.ExecuteCommand(updatedArgs)
	
	// Update LastResult
	env.LastResult = env.Driver.LastResult()
	
	return err
}