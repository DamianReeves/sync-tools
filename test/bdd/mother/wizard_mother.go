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

// Advanced scenario builders using composition and domain-specific patterns

// DevOpsScenarios provides pre-built configurations for DevOps use cases
type DevOpsScenarios struct{}

func (d DevOpsScenarios) DeploymentSync(sourceDir, destDir string) *wizard.TestModeOptions {
	return NewWizardConfig().
		WithSourceDir(sourceDir).
		WithDestinationDir(destDir).
		WithMode("one-way").
		WithExclusionPatterns("*.tmp", "*.log", ".git/", "node_modules/", "target/").
		WithGitIgnore(true).
		Build()
}

func (d DevOpsScenarios) BackupSync(sourceDir, destDir string) *wizard.TestModeOptions {
	return NewWizardConfig().
		WithSourceDir(sourceDir).
		WithDestinationDir(destDir).
		WithMode("two-way").
		WithExclusionPatterns("*.cache", "*.lock").
		WithGitIgnore(false).
		Build()
}

func (d DevOpsScenarios) ConfigurationSync(sourceDir, destDir string) *wizard.TestModeOptions {
	return NewWizardConfig().
		WithSourceDir(sourceDir).
		WithDestinationDir(destDir).
		WithMode("two-way").
		WithExclusionPatterns("*.bak", "*.swp").
		WithDryRun(true). // Always dry-run for config changes
		Build()
}

// DeveloperScenarios provides pre-built configurations for developer workflows
type DeveloperScenarios struct{}

func (d DeveloperScenarios) ProjectSync(sourceDir, destDir string) *wizard.TestModeOptions {
	return NewWizardConfig().
		WithSourceDir(sourceDir).
		WithDestinationDir(destDir).
		WithMode("one-way").
		WithExclusionPatterns("node_modules/", "target/", "build/", "dist/", "*.o", "*.class").
		WithGitIgnore(true).
		Build()
}

func (d DeveloperScenarios) AssetSync(sourceDir, destDir string) *wizard.TestModeOptions {
	return NewWizardConfig().
		WithSourceDir(sourceDir).
		WithDestinationDir(destDir).
		WithMode("two-way").
		WithExclusionPatterns("*.tmp").
		Build()
}

func (d DeveloperScenarios) DocumentationSync(sourceDir, destDir string) *wizard.TestModeOptions {
	return NewWizardConfig().
		WithSourceDir(sourceDir).
		WithDestinationDir(destDir).
		WithMode("one-way").
		WithExclusionPatterns(".DS_Store", "Thumbs.db").
		Build()
}

// ComplianceScenarios provides pre-built configurations for audit and compliance
type ComplianceScenarios struct{}

func (c ComplianceScenarios) AuditSync(sourceDir, destDir string) *wizard.TestModeOptions {
	return NewWizardConfig().
		WithSourceDir(sourceDir).
		WithDestinationDir(destDir).
		WithMode("one-way").
		WithDryRun(true).     // Compliance always requires dry-run first
		WithGitIgnore(false). // Include all files for compliance
		Build()
}

func (c ComplianceScenarios) ArchivalSync(sourceDir, destDir string) *wizard.TestModeOptions {
	return NewWizardConfig().
		WithSourceDir(sourceDir).
		WithDestinationDir(destDir).
		WithMode("one-way").
		WithGitIgnore(false).
		Build()
}

// Complex multi-stage scenario builders

// MultiStageWizardBuilder builds scenarios that require multiple wizard executions
type MultiStageWizardBuilder interface {
	Stage1() WizardConfigBuilder
	Stage2(previousResult *wizard.TestModeOptions) WizardConfigBuilder
	Stage3(stage1Result, stage2Result *wizard.TestModeOptions) WizardConfigBuilder
	Build() []*wizard.TestModeOptions
}

type multiStageBuilder struct {
	stages []*wizard.TestModeOptions
}

func NewMultiStageWizard() MultiStageWizardBuilder {
	return &multiStageBuilder{
		stages: make([]*wizard.TestModeOptions, 0),
	}
}

func (m *multiStageBuilder) Stage1() WizardConfigBuilder {
	return &stageConfigBuilder{parent: m, stageIndex: 0}
}

func (m *multiStageBuilder) Stage2(previousResult *wizard.TestModeOptions) WizardConfigBuilder {
	return &stageConfigBuilder{parent: m, stageIndex: 1, previous: previousResult}
}

func (m *multiStageBuilder) Stage3(stage1Result, stage2Result *wizard.TestModeOptions) WizardConfigBuilder {
	return &stageConfigBuilder{parent: m, stageIndex: 2}
}

func (m *multiStageBuilder) Build() []*wizard.TestModeOptions {
	return m.stages
}

type stageConfigBuilder struct {
	parent     *multiStageBuilder
	stageIndex int
	previous   *wizard.TestModeOptions
	current    *wizardConfigBuilder
}

func (s *stageConfigBuilder) WithSourceDir(source string) WizardConfigBuilder {
	if s.current == nil {
		s.current = NewWizardConfig().(*wizardConfigBuilder)
	}
	return s.current.WithSourceDir(source)
}

func (s *stageConfigBuilder) WithDestinationDir(dest string) WizardConfigBuilder {
	if s.current == nil {
		s.current = NewWizardConfig().(*wizardConfigBuilder)
	}
	return s.current.WithDestinationDir(dest)
}

func (s *stageConfigBuilder) WithMode(mode string) WizardConfigBuilder {
	if s.current == nil {
		s.current = NewWizardConfig().(*wizardConfigBuilder)
	}
	return s.current.WithMode(mode)
}

func (s *stageConfigBuilder) WithExclusionPattern(pattern string) WizardConfigBuilder {
	if s.current == nil {
		s.current = NewWizardConfig().(*wizardConfigBuilder)
	}
	return s.current.WithExclusionPattern(pattern)
}

func (s *stageConfigBuilder) WithExclusionPatterns(patterns ...string) WizardConfigBuilder {
	if s.current == nil {
		s.current = NewWizardConfig().(*wizardConfigBuilder)
	}
	return s.current.WithExclusionPatterns(patterns...)
}

func (s *stageConfigBuilder) WithGitIgnore(enabled bool) WizardConfigBuilder {
	if s.current == nil {
		s.current = NewWizardConfig().(*wizardConfigBuilder)
	}
	return s.current.WithGitIgnore(enabled)
}

func (s *stageConfigBuilder) WithDryRun(enabled bool) WizardConfigBuilder {
	if s.current == nil {
		s.current = NewWizardConfig().(*wizardConfigBuilder)
	}
	return s.current.WithDryRun(enabled)
}

func (s *stageConfigBuilder) Build() *wizard.TestModeOptions {
	if s.current == nil {
		s.current = NewWizardConfig().(*wizardConfigBuilder)
	}
	result := s.current.Build()

	// Add to parent's stages
	if len(s.parent.stages) <= s.stageIndex {
		// Extend slice if needed
		for len(s.parent.stages) <= s.stageIndex {
			s.parent.stages = append(s.parent.stages, nil)
		}
	}
	s.parent.stages[s.stageIndex] = result

	return result
}

// Persona-specific wizard mothers that embody domain knowledge

// ProductManagerWizard creates wizard configurations from a PM perspective
func ProductManagerWizard() ProductManagerWizardBuilder {
	return &productManagerWizardBuilder{}
}

type ProductManagerWizardBuilder interface {
	FeatureDeployment(sourceDir, destDir string) *wizard.TestModeOptions
	HotfixDeployment(sourceDir, destDir string) *wizard.TestModeOptions
	RollbackScenario(sourceDir, destDir string) *wizard.TestModeOptions
}

type productManagerWizardBuilder struct{}

func (p *productManagerWizardBuilder) FeatureDeployment(sourceDir, destDir string) *wizard.TestModeOptions {
	return NewWizardConfig().
		WithSourceDir(sourceDir).
		WithDestinationDir(destDir).
		WithMode("one-way").
		WithExclusionPatterns("*.test.*", "*.spec.*", "test/", "spec/").
		WithDryRun(true). // PM always wants to review first
		Build()
}

func (p *productManagerWizardBuilder) HotfixDeployment(sourceDir, destDir string) *wizard.TestModeOptions {
	return NewWizardConfig().
		WithSourceDir(sourceDir).
		WithDestinationDir(destDir).
		WithMode("one-way").
		WithDryRun(false). // Hotfixes need to be fast
		Build()
}

func (p *productManagerWizardBuilder) RollbackScenario(sourceDir, destDir string) *wizard.TestModeOptions {
	return NewWizardConfig().
		WithSourceDir(sourceDir).
		WithDestinationDir(destDir).
		WithMode("one-way").
		WithDryRun(true). // Rollbacks require careful review
		Build()
}

// ComplianceAuditorWizard creates wizard configurations from a compliance perspective
func ComplianceAuditorWizard() ComplianceAuditorWizardBuilder {
	return &complianceAuditorWizardBuilder{}
}

type ComplianceAuditorWizardBuilder interface {
	DataRetentionSync(sourceDir, destDir string) *wizard.TestModeOptions
	SecurityAuditSync(sourceDir, destDir string) *wizard.TestModeOptions
	ComplianceReportSync(sourceDir, destDir string) *wizard.TestModeOptions
}

type complianceAuditorWizardBuilder struct{}

func (c *complianceAuditorWizardBuilder) DataRetentionSync(sourceDir, destDir string) *wizard.TestModeOptions {
	return NewWizardConfig().
		WithSourceDir(sourceDir).
		WithDestinationDir(destDir).
		WithMode("one-way").
		WithGitIgnore(false). // Include all files for compliance
		WithDryRun(true).     // Always dry-run for compliance
		Build()
}

func (c *complianceAuditorWizardBuilder) SecurityAuditSync(sourceDir, destDir string) *wizard.TestModeOptions {
	return NewWizardConfig().
		WithSourceDir(sourceDir).
		WithDestinationDir(destDir).
		WithMode("one-way").
		WithGitIgnore(false). // Include all files for security audit
		Build()
}

func (c *complianceAuditorWizardBuilder) ComplianceReportSync(sourceDir, destDir string) *wizard.TestModeOptions {
	return NewWizardConfig().
		WithSourceDir(sourceDir).
		WithDestinationDir(destDir).
		WithMode("one-way").
		WithExclusionPatterns("*.tmp", "*.cache").
		Build()
}

// Convenience functions for common patterns
var (
	DevOps     = DevOpsScenarios{}
	Developer  = DeveloperScenarios{}
	Compliance = ComplianceScenarios{}
)
