package mother

import (
	"github.com/DamianReeves/sync-tools/internal/wizard"
)

// WizardConfigBuilder provides a fluent API for creating wizard test configurations
type WizardConfigBuilder interface {
	WithSourceDir(source string) WizardConfigBuilder
	WithDestinationDir(dest string) WizardConfigBuilder
	WithMode(mode string) WizardConfigBuilder
	WithExclusionPattern(pattern string) WizardConfigBuilder
	WithExclusionPatterns(patterns ...string) WizardConfigBuilder
	WithGitIgnore(enabled bool) WizardConfigBuilder
	WithDryRun(enabled bool) WizardConfigBuilder
	Build() *wizard.TestModeOptions
}

type wizardConfigBuilder struct {
	sourceDir         string
	destinationDir    string
	mode              string
	exclusionPatterns []string
	enableGitIgnore   bool
	dryRun            bool
}

// NewWizardConfig creates a new wizard configuration builder
func NewWizardConfig() WizardConfigBuilder {
	return &wizardConfigBuilder{
		mode:              "one-way", // default
		exclusionPatterns: make([]string, 0),
		enableGitIgnore:   false,
		dryRun:            false,
	}
}

func (b *wizardConfigBuilder) WithSourceDir(source string) WizardConfigBuilder {
	b.sourceDir = source
	return b
}

func (b *wizardConfigBuilder) WithDestinationDir(dest string) WizardConfigBuilder {
	b.destinationDir = dest
	return b
}

func (b *wizardConfigBuilder) WithMode(mode string) WizardConfigBuilder {
	b.mode = mode
	return b
}

func (b *wizardConfigBuilder) WithExclusionPattern(pattern string) WizardConfigBuilder {
	b.exclusionPatterns = append(b.exclusionPatterns, pattern)
	return b
}

func (b *wizardConfigBuilder) WithExclusionPatterns(patterns ...string) WizardConfigBuilder {
	b.exclusionPatterns = append(b.exclusionPatterns, patterns...)
	return b
}

func (b *wizardConfigBuilder) WithGitIgnore(enabled bool) WizardConfigBuilder {
	b.enableGitIgnore = enabled
	return b
}

func (b *wizardConfigBuilder) WithDryRun(enabled bool) WizardConfigBuilder {
	b.dryRun = enabled
	return b
}

func (b *wizardConfigBuilder) Build() *wizard.TestModeOptions {
	return &wizard.TestModeOptions{
		SourceDir:         b.sourceDir,
		DestinationDir:    b.destinationDir,
		Mode:              b.mode,
		ExclusionPatterns: b.exclusionPatterns,
		EnableGitIgnore:   b.enableGitIgnore,
		DryRun:            b.dryRun,
	}
}

// Common wizard test scenarios

// BasicOneWayWizardConfig creates a basic one-way sync wizard configuration
func BasicOneWayWizardConfig(sourceDir, destDir string) *wizard.TestModeOptions {
	return NewWizardConfig().
		WithSourceDir(sourceDir).
		WithDestinationDir(destDir).
		WithMode("one-way").
		Build()
}

// TwoWaySyncWizardConfig creates a two-way sync wizard configuration
func TwoWaySyncWizardConfig(sourceDir, destDir string) *wizard.TestModeOptions {
	return NewWizardConfig().
		WithSourceDir(sourceDir).
		WithDestinationDir(destDir).
		WithMode("two-way").
		Build()
}

// WizardWithExclusionPatterns creates a wizard configuration with exclusion patterns
func WizardWithExclusionPatterns(sourceDir, destDir string, patterns ...string) *wizard.TestModeOptions {
	return NewWizardConfig().
		WithSourceDir(sourceDir).
		WithDestinationDir(destDir).
		WithExclusionPatterns(patterns...).
		Build()
}

// WizardWithGitIgnore creates a wizard configuration with git ignore enabled
func WizardWithGitIgnore(sourceDir, destDir string) *wizard.TestModeOptions {
	return NewWizardConfig().
		WithSourceDir(sourceDir).
		WithDestinationDir(destDir).
		WithGitIgnore(true).
		Build()
}

// ComplexWizardConfig creates a complex wizard configuration with multiple options
func ComplexWizardConfig(sourceDir, destDir string) *wizard.TestModeOptions {
	return NewWizardConfig().
		WithSourceDir(sourceDir).
		WithDestinationDir(destDir).
		WithMode("two-way").
		WithExclusionPatterns("*.tmp", "*.log", "node_modules/").
		WithGitIgnore(true).
		WithDryRun(false).
		Build()
}

// DryRunWizardConfig creates a dry-run wizard configuration
func DryRunWizardConfig(sourceDir, destDir string) *wizard.TestModeOptions {
	return NewWizardConfig().
		WithSourceDir(sourceDir).
		WithDestinationDir(destDir).
		WithDryRun(true).
		Build()
}

// WizardConfigBuilder allows for more complex scenarios
func WizardConfigFor(sourceDir, destDir string) WizardConfigBuilder {
	return NewWizardConfig().
		WithSourceDir(sourceDir).
		WithDestinationDir(destDir)
}
