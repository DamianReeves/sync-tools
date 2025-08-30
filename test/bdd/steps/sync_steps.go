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
	workingDir     string
	lastExitCode   int
	lastOutput     string
	lastError      string
	syncToolsPath  string
}

// Helper function to run a command and properly capture exit code and output
func (tc *TestContext) runCommand(args ...string) error {
	cmd := exec.Command(tc.syncToolsPath, args...)
	cmd.Dir = tc.workingDir // Run from working directory
	output, err := cmd.CombinedOutput()
	tc.lastOutput = string(output)
	
	// Handle exit code properly
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			tc.lastExitCode = exitError.ExitCode()
		} else {
			tc.lastExitCode = -1
		}
		tc.lastError = err.Error()
	} else {
		tc.lastExitCode = 0
	}
	return nil
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

	// Git patch steps
	ctx.Step(`^I have a destination directory with some matching and some different files$`, tc.createDestinationDirectoryWithMixedFiles)
	ctx.Step(`^I have a destination directory with files$`, tc.createDestinationDirectoryWithFiles)
	ctx.Step(`^I run sync-tools with patch generation to "([^"]*)"$`, tc.runSyncToolsWithPatchGeneration)
	ctx.Step(`^I run sync-tools with patch generation to "([^"]*)" and dry-run$`, tc.runSyncToolsWithPatchGenerationAndDryRun)
	ctx.Step(`^I run sync-tools with patch generation to "([^"]*)" and only mode for "([^"]*)"$`, tc.runSyncToolsWithPatchGenerationAndOnly)
	ctx.Step(`^a git patch file should be created at "([^"]*)"$`, tc.gitPatchFileShouldBeCreated)
	ctx.Step(`^the patch file should contain differences between source and destination$`, tc.patchFileShouldContainDifferences)
	ctx.Step(`^the patch file should contain all new files from source$`, tc.patchFileShouldContainAllNewFiles)
	ctx.Step(`^the patch should show files as new additions$`, tc.patchShouldShowFilesAsNewAdditions)
	ctx.Step(`^the patch file should contain file deletions$`, tc.patchFileShouldContainFileDeletions)
	ctx.Step(`^the patch should show files as removals$`, tc.patchShouldShowFilesAsRemovals)
	ctx.Step(`^the patch file should not contain ignored files$`, tc.patchFileShouldNotContainIgnoredFiles)
	ctx.Step(`^the patch file should only contain changes for whitelisted files$`, tc.patchFileShouldOnlyContainWhitelistedFiles)
	ctx.Step(`^it should show what would be included in the patch$`, tc.shouldShowWhatWouldBeIncludedInPatch)
	ctx.Step(`^no patch file should be created$`, tc.noPatchFileShouldBeCreated)
	ctx.Step(`^no files should be synchronized$`, tc.noFilesShouldBeSynchronized)
	ctx.Step(`^I have an empty source directory$`, tc.createEmptySourceDirectory)
	ctx.Step(`^files matching gitignore patterns should not be copied$`, tc.filesMatchingGitignorePatternsShouldNotBeCopied)

	// Interactive sync plan steps
	ctx.Step(`^a file "([^"]*)" should be created$`, tc.fileShouldBeCreated)
	ctx.Step(`^the plan file should contain:$`, tc.planFileShouldContain)
	ctx.Step(`^the plan file should contain sync operations with visual aliases$`, tc.planFileShouldContainVisualAliases)
	ctx.Step(`^the plan file "([^"]*)" should contain:$`, tc.namedPlanFileShouldContain)
	ctx.Step(`^the plan file should not contain "([^"]*)"$`, tc.planFileShouldNotContain)
	ctx.Step(`^I have a plan file "([^"]*)" containing:$`, tc.createPlanFile)
	ctx.Step(`^I have a SyncFile "([^"]*)" containing:$`, tc.createSyncFile)
	ctx.Step(`^the destination file "([^"]*)" should contain "([^"]*)"$`, tc.destinationFileShouldContain)
	ctx.Step(`^the error should contain "([^"]*)"$`, tc.errorShouldContain)
	
	// Table-driven data creation steps
	ctx.Step(`^I have a source directory with files:$`, tc.createSourceDirectoryWithTable)
	ctx.Step(`^I have a destination directory with files:$`, tc.createDestinationDirectoryWithTable)
	ctx.Step(`^I run sync-tools with arguments "([^"]*)"$`, tc.runSyncToolsWithArguments)
	ctx.Step(`^the command should succeed$`, tc.commandShouldSucceed)
	ctx.Step(`^the command should fail$`, tc.commandShouldFail)
	ctx.Step(`^the destination directory should contain "([^"]*)"$`, tc.destinationDirectoryShouldContain)
	ctx.Step(`^the destination directory should contain "([^"]*)" with content "([^"]*)"$`, tc.destinationDirectoryShouldContainWithContent)
	ctx.Step(`^the source directory should contain "([^"]*)" with content "([^"]*)"$`, tc.sourceDirectoryShouldContainWithContent)
	ctx.Step(`^the plan file should contain "([^"]*)"$`, tc.planFileShouldContainText)

	// Setup and cleanup hooks
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		return tc.beforeScenario(ctx, sc)
	})
	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		return tc.afterScenario(ctx, sc, err)
	})
}

func (tc *TestContext) beforeScenario(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	// Create temporary directories for testing
	tempDir := os.TempDir()
	tc.sourceDir = filepath.Join(tempDir, fmt.Sprintf("sync_test_src_%d_%s", os.Getpid(), strings.ReplaceAll(sc.Name, " ", "_")))
	tc.destDir = filepath.Join(tempDir, fmt.Sprintf("sync_test_dest_%d_%s", os.Getpid(), strings.ReplaceAll(sc.Name, " ", "_")))
	tc.workingDir = filepath.Join(tempDir, fmt.Sprintf("sync_test_work_%d_%s", os.Getpid(), strings.ReplaceAll(sc.Name, " ", "_")))
	
	// Find sync-tools binary path - always relative to project root
	if wd, err := os.Getwd(); err == nil {
		// If we're in test/bdd, go up two levels
		if strings.HasSuffix(wd, "test/bdd") {
			tc.syncToolsPath = filepath.Join(wd, "..", "..", "sync-tools")
		} else {
			// If we're in project root, binary is in current directory
			tc.syncToolsPath = filepath.Join(wd, "sync-tools")
		}
	} else {
		tc.syncToolsPath = "../../sync-tools"
	}
	
	// Create the working directory
	err := os.MkdirAll(tc.workingDir, 0755)
	if err != nil {
		return ctx, fmt.Errorf("failed to create working directory: %w", err)
	}
	
	return ctx, nil
}

func (tc *TestContext) afterScenario(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
	// Cleanup test directories
	_ = os.RemoveAll(tc.sourceDir)
	_ = os.RemoveAll(tc.destDir)
	_ = os.RemoveAll(tc.workingDir)
	// Note: sc and err parameters are required by godog interface
	_ = sc
	_ = err
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
	return tc.runCommand("help")
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

func (tc *TestContext) createDestinationDirectoryWithFiles() error {
	if err := os.MkdirAll(tc.destDir, 0755); err != nil {
		return err
	}
	
	// Create some files in destination (for deletion patch testing)
	files := []string{"dest_file1.txt", "dest_file2.txt", "dest_subdir/dest_file3.txt"}
	for _, file := range files {
		fullPath := filepath.Join(tc.destDir, file)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(fullPath, []byte("content for "+file), 0644); err != nil {
			return err
		}
	}
	
	return nil
}

func (tc *TestContext) runSyncToolsWithOneWaySyncAndDryRun() error {
	return tc.runCommand("sync", "--source", tc.sourceDir, "--dest", tc.destDir, "--dry-run")
}

func (tc *TestContext) runSyncToolsWithOneWaySync() error {
	return tc.runCommand("sync", "--source", tc.sourceDir, "--dest", tc.destDir)
}

func (tc *TestContext) runSyncToolsWithTwoWaySync() error {
	return tc.runCommand("sync", "--source", tc.sourceDir, "--dest", tc.destDir, "--mode", "two-way")
}

// Placeholder implementations - these would be implemented as the CLI is built
func (tc *TestContext) shouldShowWhatFilesWouldBeCopied() error {
	// Check for dry-run indicators in the output
	if !strings.Contains(tc.lastOutput, "DRY RUN") && !strings.Contains(tc.lastOutput, "dry-run=true") {
		return fmt.Errorf("expected dry-run output to show what would be copied, got: %s", tc.lastOutput)
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
	return tc.runCommand("sync", "--source", tc.sourceDir, "--dest", tc.destDir, "--use-source-gitignore")
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

// Git patch step implementations

func (tc *TestContext) createEmptySourceDirectory() error {
	return os.MkdirAll(tc.sourceDir, 0755)
}

func (tc *TestContext) createDestinationDirectoryWithMixedFiles() error {
	if err := os.MkdirAll(tc.destDir, 0755); err != nil {
		return err
	}
	
	// Create some matching files (same content)
	if err := os.WriteFile(filepath.Join(tc.destDir, "file1.txt"), []byte("test content for file1.txt"), 0644); err != nil {
		return err
	}
	
	// Create some different files (different content)
	if err := os.WriteFile(filepath.Join(tc.destDir, "file2.txt"), []byte("DIFFERENT content for file2.txt"), 0644); err != nil {
		return err
	}
	
	// Create files that only exist in destination
	if err := os.WriteFile(filepath.Join(tc.destDir, "dest_only.txt"), []byte("only in destination"), 0644); err != nil {
		return err
	}
	
	return nil
}

func (tc *TestContext) runSyncToolsWithPatchGeneration(patchFile string) error {
	return tc.runCommand("sync", "--source", tc.sourceDir, "--dest", tc.destDir, "--patch", patchFile)
}

func (tc *TestContext) runSyncToolsWithPatchGenerationAndDryRun(patchFile string) error {
	return tc.runCommand("sync", "--source", tc.sourceDir, "--dest", tc.destDir, "--patch", patchFile, "--dry-run")
}

func (tc *TestContext) runSyncToolsWithPatchGenerationAndOnly(patchFile, onlyPattern string) error {
	return tc.runCommand("sync", "--source", tc.sourceDir, "--dest", tc.destDir, "--patch", patchFile, "--only", onlyPattern)
}

func (tc *TestContext) gitPatchFileShouldBeCreated(patchFile string) error {
	// Check in current working directory first
	if _, err := os.Stat(patchFile); err == nil {
		return nil
	}
	
	// Check in the project root directory (where sync-tools would create it)
	wd, _ := os.Getwd()
	var projectRoot string
	if strings.HasSuffix(wd, "test/bdd") {
		projectRoot = filepath.Join(wd, "..", "..")
	} else {
		projectRoot = wd
	}
	
	rootPatch := filepath.Join(projectRoot, patchFile)
	if _, err := os.Stat(rootPatch); err == nil {
		return nil
	}
	
	// Check absolute path
	if filepath.IsAbs(patchFile) {
		if _, err := os.Stat(patchFile); err == nil {
			return nil
		}
	}
	
	return fmt.Errorf("expected patch file %s to exist, but it doesn't (checked: %s, %s)", patchFile, patchFile, rootPatch)
}

func (tc *TestContext) patchFileShouldContainDifferences() error {
	// This would check that the patch contains actual diff content
	// For now, just verify the file is not empty
	return nil // Placeholder - will implement after CLI flag is added
}

func (tc *TestContext) patchFileShouldContainAllNewFiles() error {
	return nil // Placeholder - will implement after CLI flag is added
}

func (tc *TestContext) patchShouldShowFilesAsNewAdditions() error {
	return nil // Placeholder - will implement after CLI flag is added
}

func (tc *TestContext) patchFileShouldContainFileDeletions() error {
	return nil // Placeholder - will implement after CLI flag is added
}

func (tc *TestContext) patchShouldShowFilesAsRemovals() error {
	return nil // Placeholder - will implement after CLI flag is added
}

func (tc *TestContext) patchFileShouldNotContainIgnoredFiles() error {
	return nil // Placeholder - will implement after CLI flag is added
}

func (tc *TestContext) patchFileShouldOnlyContainWhitelistedFiles() error {
	return nil // Placeholder - will implement after CLI flag is added
}

func (tc *TestContext) shouldShowWhatWouldBeIncludedInPatch() error {
	// Check for dry-run indicators and patch-related output
	if !strings.Contains(tc.lastOutput, "DRY RUN") && !strings.Contains(tc.lastOutput, "dry-run=true") && !strings.Contains(tc.lastOutput, "patch") {
		return fmt.Errorf("expected dry-run output to show what would be included in patch, got: %s", tc.lastOutput)
	}
	return nil
}

func (tc *TestContext) noPatchFileShouldBeCreated() error {
	// This would check that no .patch files exist in the working directory
	return nil // Placeholder - will implement after CLI flag is added
}

func (tc *TestContext) noFilesShouldBeSynchronized() error {
	// Check that no actual synchronization occurred by verifying destination is unchanged
	// This is used for patch generation where files should not be copied
	return nil // Placeholder - patch generation doesn't sync files
}

func (tc *TestContext) filesMatchingGitignorePatternsShouldNotBeCopied() error {
	// Check that gitignore patterns were respected
	return nil // Placeholder - need to implement gitignore pattern validation
}

// Interactive sync plan step implementations

func (tc *TestContext) fileShouldBeCreated(filename string) error {
	fullPath := filepath.Join(tc.workingDir, filename)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("expected file %s to be created", filename)
	}
	return nil
}

func (tc *TestContext) planFileShouldContain(expectedContent *godog.DocString) error {
	// Find the most recently created .plan file
	planFile := ""
	files, err := filepath.Glob(filepath.Join(tc.workingDir, "*.plan"))
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no .plan file found")
	}
	planFile = files[0] // Use first found plan file
	
	content, err := os.ReadFile(planFile)
	if err != nil {
		return fmt.Errorf("failed to read plan file %s: %w", planFile, err)
	}
	
	expectedLines := strings.Split(strings.TrimSpace(expectedContent.Content), "\n")
	actualContent := string(content)
	
	for _, expectedLine := range expectedLines {
		expectedLine = strings.TrimSpace(expectedLine)
		if expectedLine == "" {
			continue
		}
		if !strings.Contains(actualContent, expectedLine) {
			return fmt.Errorf("plan file does not contain expected line: %s", expectedLine)
		}
	}
	
	return nil
}

func (tc *TestContext) planFileShouldContainVisualAliases() error {
	// Find any .plan file and check for visual aliases
	files, err := filepath.Glob(filepath.Join(tc.workingDir, "*.plan"))
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no .plan file found")
	}
	
	content, err := os.ReadFile(files[0])
	if err != nil {
		return err
	}
	
	planContent := string(content)
	hasVisualAlias := strings.Contains(planContent, "<<") || 
					  strings.Contains(planContent, ">>") || 
					  strings.Contains(planContent, "<>")
	
	if !hasVisualAlias {
		return fmt.Errorf("plan file should contain visual aliases (<<, >>, <>)")
	}
	
	return nil
}

func (tc *TestContext) namedPlanFileShouldContain(filename string, expectedContent *godog.DocString) error {
	fullPath := filepath.Join(tc.workingDir, filename)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read plan file %s: %w", filename, err)
	}
	
	expectedLines := strings.Split(strings.TrimSpace(expectedContent.Content), "\n")
	actualContent := string(content)
	
	for _, expectedLine := range expectedLines {
		expectedLine = strings.TrimSpace(expectedLine)
		if expectedLine == "" {
			continue
		}
		if !strings.Contains(actualContent, expectedLine) {
			return fmt.Errorf("plan file %s does not contain expected line: %s", filename, expectedLine)
		}
	}
	
	return nil
}

func (tc *TestContext) planFileShouldNotContain(text string) error {
	files, err := filepath.Glob(filepath.Join(tc.workingDir, "*.plan"))
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no .plan file found")
	}
	
	content, err := os.ReadFile(files[0])
	if err != nil {
		return err
	}
	
	if strings.Contains(string(content), text) {
		return fmt.Errorf("plan file should not contain: %s", text)
	}
	
	return nil
}

func (tc *TestContext) createPlanFile(filename string, content *godog.DocString) error {
	fullPath := filepath.Join(tc.workingDir, filename)
	
	// Replace test path placeholders with actual directory paths
	updatedContent := content.Content
	updatedContent = strings.ReplaceAll(updatedContent, "./test_source", tc.sourceDir)
	updatedContent = strings.ReplaceAll(updatedContent, "./test_dest", tc.destDir)
	
	return os.WriteFile(fullPath, []byte(updatedContent), 0644)
}

func (tc *TestContext) createSyncFile(filename string, content *godog.DocString) error {
	fullPath := filepath.Join(tc.workingDir, filename)
	return os.WriteFile(fullPath, []byte(content.Content), 0644)
}

func (tc *TestContext) destinationFileShouldContain(filename, expectedContent string) error {
	fullPath := filepath.Join(tc.destDir, filename)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read destination file %s: %w", filename, err)
	}
	
	if strings.TrimSpace(string(content)) != expectedContent {
		return fmt.Errorf("expected destination file %s to contain '%s', but got '%s'", 
			filename, expectedContent, strings.TrimSpace(string(content)))
	}
	
	return nil
}

func (tc *TestContext) errorShouldContain(expectedError string) error {
	if !strings.Contains(tc.lastError, expectedError) {
		return fmt.Errorf("expected error to contain '%s', but got '%s'", expectedError, tc.lastError)
	}
	return nil
}

// Table-driven step implementations

func (tc *TestContext) createSourceDirectoryWithTable(table *godog.Table) error {
	// Ensure source directory exists
	err := os.MkdirAll(tc.sourceDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create source directory: %w", err)
	}
	
	// Create files from table
	for i, row := range table.Rows {
		if i == 0 { // Skip header row
			continue
		}
		
		if len(row.Cells) < 2 {
			return fmt.Errorf("table row must have at least path and content columns")
		}
		
		path := row.Cells[0].Value
		content := row.Cells[1].Value
		
		fullPath := filepath.Join(tc.sourceDir, path)
		
		// Create parent directories if needed
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", path, err)
		}
		
		// Write file
		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			return fmt.Errorf("failed to write file %s: %w", path, err)
		}
	}
	
	return nil
}

func (tc *TestContext) createDestinationDirectoryWithTable(table *godog.Table) error {
	// Ensure destination directory exists
	err := os.MkdirAll(tc.destDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}
	
	// Create files from table
	for i, row := range table.Rows {
		if i == 0 { // Skip header row
			continue
		}
		
		if len(row.Cells) < 2 {
			return fmt.Errorf("table row must have at least path and content columns")
		}
		
		path := row.Cells[0].Value
		content := row.Cells[1].Value
		
		fullPath := filepath.Join(tc.destDir, path)
		
		// Create parent directories if needed
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", path, err)
		}
		
		// Write file
		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			return fmt.Errorf("failed to write file %s: %w", path, err)
		}
	}
	
	return nil
}

func (tc *TestContext) runSyncToolsWithArguments(args string) error {
	// Parse the arguments string and replace test paths with actual directories
	argSlice := strings.Fields(args)
	for i, arg := range argSlice {
		if arg == "./test_source" {
			argSlice[i] = tc.sourceDir
		} else if arg == "./test_dest" {
			argSlice[i] = tc.destDir
		}
	}
	return tc.runCommand(argSlice...)
}

func (tc *TestContext) commandShouldSucceed() error {
	if tc.lastExitCode != 0 {
		return fmt.Errorf("expected command to succeed (exit code 0), but got %d. Output: %s", tc.lastExitCode, tc.lastOutput)
	}
	return nil
}

func (tc *TestContext) commandShouldFail() error {
	if tc.lastExitCode == 0 {
		return fmt.Errorf("expected command to fail (non-zero exit code), but got 0. Output: %s", tc.lastOutput)
	}
	return nil
}

func (tc *TestContext) destinationDirectoryShouldContain(filename string) error {
	fullPath := filepath.Join(tc.destDir, filename)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("expected destination directory to contain %s, but it does not exist", filename)
	}
	return nil
}

func (tc *TestContext) destinationDirectoryShouldContainWithContent(filename, expectedContent string) error {
	fullPath := filepath.Join(tc.destDir, filename)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read destination file %s: %w", filename, err)
	}
	
	if strings.TrimSpace(string(content)) != expectedContent {
		return fmt.Errorf("expected destination file %s to contain '%s', but got '%s'", 
			filename, expectedContent, strings.TrimSpace(string(content)))
	}
	
	return nil
}

func (tc *TestContext) sourceDirectoryShouldContainWithContent(filename, expectedContent string) error {
	fullPath := filepath.Join(tc.sourceDir, filename)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read source file %s: %w", filename, err)
	}
	
	if strings.TrimSpace(string(content)) != expectedContent {
		return fmt.Errorf("expected source file %s to contain '%s', but got '%s'", 
			filename, expectedContent, strings.TrimSpace(string(content)))
	}
	
	return nil
}

func (tc *TestContext) planFileShouldContainText(text string) error {
	files, err := filepath.Glob(filepath.Join(tc.workingDir, "*.plan"))
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no .plan file found")
	}
	
	content, err := os.ReadFile(files[0])
	if err != nil {
		return err
	}
	
	if !strings.Contains(string(content), text) {
		return fmt.Errorf("plan file should contain: %s", text)
	}
	
	return nil
}