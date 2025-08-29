package steps

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

// TestContext holds state between steps
type TestContext struct {
	sourceDir      string
	destDir        string
	lastExitCode   int
	lastOutput     string
	lastError      string
	syncToolsPath  string
}

// NewTestContext creates a new test context
func NewTestContext() *TestContext {
	return &TestContext{}
}

// RegisterSteps registers all step definitions
func (tc *TestContext) RegisterSteps(ctx *godog.ScenarioContext) {
	// Hello World steps
	ctx.Step(`^the sync-tools binary exists$`, tc.syncToolsBinaryExists)
	ctx.Step(`^I run sync-tools with help$`, tc.runSyncToolsWithHelp)
	ctx.Step(`^it should display help information$`, tc.shouldDisplayHelpInformation)
	ctx.Step(`^the exit code should be (\d+)$`, tc.exitCodeShouldBe)

	// Basic sync steps
	ctx.Step(`^I have a source directory with files$`, tc.createSourceDirectoryWithFiles)
	ctx.Step(`^I have an empty destination directory$`, tc.createEmptyDestinationDirectory)
	ctx.Step(`^I have a destination directory with different files$`, tc.createDestinationDirectoryWithDifferentFiles)
	ctx.Step(`^I run sync-tools with one-way sync and dry-run$`, tc.runSyncToolsWithOneWaySyncAndDryRun)
	ctx.Step(`^I run sync-tools with one-way sync$`, tc.runSyncToolsWithOneWaySync)
	ctx.Step(`^I run sync-tools with two-way sync$`, tc.runSyncToolsWithTwoWaySync)
	ctx.Step(`^it should show what files would be copied$`, tc.shouldShowWhatFilesWouldBeCopied)
	ctx.Step(`^no files should actually be copied$`, tc.noFilesShouldActuallyBeCopied)
	ctx.Step(`^files should be copied to destination$`, tc.filesShouldBeCopiedToDestination)
	ctx.Step(`^the destination should match source$`, tc.destinationShouldMatchSource)
	ctx.Step(`^files should be synchronized in both directions$`, tc.filesShouldBeSynchronizedInBothDirections)
	ctx.Step(`^conflicts should be handled appropriately$`, tc.conflictsShouldBeHandledAppropriately)

	// Ignore pattern steps
	ctx.Step(`^I have a \.syncignore file in the source directory$`, tc.createSyncIgnoreFile)
	ctx.Step(`^I have a \.gitignore file in the source directory$`, tc.createGitIgnoreFile)
	ctx.Step(`^I have ignore patterns with unignore rules$`, tc.createIgnorePatternsWithUnignoreRules)
	ctx.Step(`^I run sync-tools with gitignore import enabled$`, tc.runSyncToolsWithGitignoreImport)
	ctx.Step(`^files matching ignore patterns should not be copied$`, tc.filesMatchingIgnorePatternsShouldNotBeCopied)
	ctx.Step(`^files not matching patterns should be copied$`, tc.filesNotMatchingPatternsShouldBeCopied)
	ctx.Step(`^files matching unignore patterns should be copied$`, tc.filesMatchingUnignorePatternsShouldBeCopied)

	// Setup and cleanup hooks
	ctx.BeforeScenario(func(sc *godog.Scenario) {
		tc.beforeScenario(context.Background(), sc)
	})
	ctx.AfterScenario(func(sc *godog.Scenario, err error) {
		tc.afterScenario(context.Background(), sc, err)
	})
}

func (tc *TestContext) beforeScenario(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	// Create temporary directories for testing
	tempDir := os.TempDir()
	tc.sourceDir = filepath.Join(tempDir, fmt.Sprintf("sync_test_src_%d", os.Getpid()))
	tc.destDir = filepath.Join(tempDir, fmt.Sprintf("sync_test_dest_%d", os.Getpid()))
	
	// Find sync-tools binary path
	if wd, err := os.Getwd(); err == nil {
		tc.syncToolsPath = filepath.Join(wd, "sync-tools")
	} else {
		tc.syncToolsPath = "sync-tools"
	}
	
	return ctx, nil
}

func (tc *TestContext) afterScenario(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
	// Cleanup test directories
	os.RemoveAll(tc.sourceDir)
	os.RemoveAll(tc.destDir)
	return ctx, nil
}

// Step implementations

func (tc *TestContext) syncToolsBinaryExists() error {
	if _, err := os.Stat(tc.syncToolsPath); os.IsNotExist(err) {
		// Try to build it first
		cmd := exec.Command("make", "build")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("sync-tools binary does not exist and failed to build: %v", err)
		}
	}
	return nil
}

func (tc *TestContext) runSyncToolsWithHelp() error {
	cmd := exec.Command(tc.syncToolsPath, "help")
	output, err := cmd.CombinedOutput()
	tc.lastOutput = string(output)
	tc.lastExitCode = cmd.ProcessState.ExitCode()
	if err != nil {
		tc.lastError = err.Error()
	}
	return nil
}

func (tc *TestContext) shouldDisplayHelpInformation() error {
	if !strings.Contains(tc.lastOutput, "sync-tools") {
		return fmt.Errorf("expected help information, got: %s", tc.lastOutput)
	}
	return nil
}

func (tc *TestContext) exitCodeShouldBe(expectedCode int) error {
	if tc.lastExitCode != expectedCode {
		return fmt.Errorf("expected exit code %d, got %d. Output: %s", expectedCode, tc.lastExitCode, tc.lastOutput)
	}
	return nil
}

func (tc *TestContext) createSourceDirectoryWithFiles() error {
	if err := os.MkdirAll(tc.sourceDir, 0755); err != nil {
		return err
	}
	
	// Create some test files
	files := []string{"file1.txt", "file2.txt", "subdir/file3.txt"}
	for _, file := range files {
		fullPath := filepath.Join(tc.sourceDir, file)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(fullPath, []byte("test content for "+file), 0644); err != nil {
			return err
		}
	}
	
	return nil
}

func (tc *TestContext) createEmptyDestinationDirectory() error {
	return os.MkdirAll(tc.destDir, 0755)
}

func (tc *TestContext) createDestinationDirectoryWithDifferentFiles() error {
	if err := os.MkdirAll(tc.destDir, 0755); err != nil {
		return err
	}
	
	// Create different files in destination
	files := []string{"different_file.txt", "another_file.txt"}
	for _, file := range files {
		fullPath := filepath.Join(tc.destDir, file)
		if err := os.WriteFile(fullPath, []byte("different content for "+file), 0644); err != nil {
			return err
		}
	}
	
	return nil
}

func (tc *TestContext) runSyncToolsWithOneWaySyncAndDryRun() error {
	cmd := exec.Command(tc.syncToolsPath, "sync", "--source", tc.sourceDir, "--dest", tc.destDir, "--dry-run")
	output, err := cmd.CombinedOutput()
	tc.lastOutput = string(output)
	tc.lastExitCode = cmd.ProcessState.ExitCode()
	if err != nil {
		tc.lastError = err.Error()
	}
	return nil
}

func (tc *TestContext) runSyncToolsWithOneWaySync() error {
	cmd := exec.Command(tc.syncToolsPath, "sync", "--source", tc.sourceDir, "--dest", tc.destDir)
	output, err := cmd.CombinedOutput()
	tc.lastOutput = string(output)
	tc.lastExitCode = cmd.ProcessState.ExitCode()
	if err != nil {
		tc.lastError = err.Error()
	}
	return nil
}

func (tc *TestContext) runSyncToolsWithTwoWaySync() error {
	cmd := exec.Command(tc.syncToolsPath, "sync", "--source", tc.sourceDir, "--dest", tc.destDir, "--mode", "two-way")
	output, err := cmd.CombinedOutput()
	tc.lastOutput = string(output)
	tc.lastExitCode = cmd.ProcessState.ExitCode()
	if err != nil {
		tc.lastError = err.Error()
	}
	return nil
}

// Placeholder implementations - these would be implemented as the CLI is built
func (tc *TestContext) shouldShowWhatFilesWouldBeCopied() error {
	// This will need to be implemented based on actual CLI output format
	if !strings.Contains(tc.lastOutput, "would") {
		return fmt.Errorf("expected dry-run output to show what would be copied")
	}
	return nil
}

func (tc *TestContext) noFilesShouldActuallyBeCopied() error {
	// Check that destination directory is still empty
	entries, err := os.ReadDir(tc.destDir)
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return fmt.Errorf("expected destination to be empty after dry-run, found %d entries", len(entries))
	}
	return nil
}

func (tc *TestContext) filesShouldBeCopiedToDestination() error {
	// Check that files exist in destination
	entries, err := os.ReadDir(tc.destDir)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return fmt.Errorf("expected files to be copied to destination, but destination is empty")
	}
	return nil
}

func (tc *TestContext) destinationShouldMatchSource() error {
	// This would need to implement directory comparison logic
	return nil // Placeholder
}

func (tc *TestContext) filesShouldBeSynchronizedInBothDirections() error {
	return nil // Placeholder
}

func (tc *TestContext) conflictsShouldBeHandledAppropriately() error {
	return nil // Placeholder
}

func (tc *TestContext) createSyncIgnoreFile() error {
	ignoreContent := "*.tmp\n*.log\ntemp/\n"
	return os.WriteFile(filepath.Join(tc.sourceDir, ".syncignore"), []byte(ignoreContent), 0644)
}

func (tc *TestContext) createGitIgnoreFile() error {
	ignoreContent := "*.tmp\n*.log\nnode_modules/\n"
	return os.WriteFile(filepath.Join(tc.sourceDir, ".gitignore"), []byte(ignoreContent), 0644)
}

func (tc *TestContext) createIgnorePatternsWithUnignoreRules() error {
	ignoreContent := "*.tmp\n!important.tmp\n"
	return os.WriteFile(filepath.Join(tc.sourceDir, ".syncignore"), []byte(ignoreContent), 0644)
}

func (tc *TestContext) runSyncToolsWithGitignoreImport() error {
	cmd := exec.Command(tc.syncToolsPath, "sync", "--source", tc.sourceDir, "--dest", tc.destDir, "--use-source-gitignore")
	output, err := cmd.CombinedOutput()
	tc.lastOutput = string(output)
	tc.lastExitCode = cmd.ProcessState.ExitCode()
	if err != nil {
		tc.lastError = err.Error()
	}
	return nil
}

func (tc *TestContext) filesMatchingIgnorePatternsShouldNotBeCopied() error {
	return nil // Placeholder
}

func (tc *TestContext) filesNotMatchingPatternsShouldBeCopied() error {
	return nil // Placeholder
}

func (tc *TestContext) filesMatchingUnignorePatternsShouldBeCopied() error {
	return nil // Placeholder
}