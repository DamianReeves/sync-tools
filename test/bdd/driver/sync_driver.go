package driver

import (
	"os/exec"
	"strings"
)

// SyncResult represents the result of a sync-tools command execution
type SyncResult struct {
	ExitCode int
	Output   string
	Error    string
	Success  bool
}

// SyncDriver provides a clean API for interacting with sync-tools in tests
type SyncDriver interface {
	// Core sync operations
	Sync(source, dest string, options ...SyncOption) *SyncResult
	GeneratePlan(source, dest, planFile string, options ...PlanOption) *SyncResult
	ApplyPlan(planFile string, options ...ApplyOption) *SyncResult
	
	// Git patch operations
	GeneratePatch(source, dest, patchFile string, options ...PatchOption) *SyncResult
	
	// Utility operations
	Help() *SyncResult
	Version() *SyncResult
	
	// Generic command execution
	ExecuteCommand(args string) error
	LastResult() *SyncResult
	
	// Configuration
	SetWorkingDir(dir string)
	SetBinaryPath(path string)
}

// syncDriver implements SyncDriver using exec.Command
type syncDriver struct {
	binaryPath string
	workingDir string
	lastResult *SyncResult
}

// NewSyncDriver creates a new sync-tools driver
func NewSyncDriver(binaryPath string) SyncDriver {
	return &syncDriver{
		binaryPath: binaryPath,
	}
}

func (d *syncDriver) SetWorkingDir(dir string) {
	d.workingDir = dir
}

func (d *syncDriver) SetBinaryPath(path string) {
	d.binaryPath = path
}

// executeCommand runs a sync-tools command and returns the result
func (d *syncDriver) executeCommand(args ...string) *SyncResult {
	cmd := exec.Command(d.binaryPath, args...)
	if d.workingDir != "" {
		cmd.Dir = d.workingDir
	}
	
	output, err := cmd.CombinedOutput()
	result := &SyncResult{
		Output: string(output),
	}
	
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = 1
			result.Error = err.Error()
		}
	} else {
		result.ExitCode = 0
		result.Success = true
	}
	
	return result
}

// Core sync operations
func (d *syncDriver) Sync(source, dest string, options ...SyncOption) *SyncResult {
	args := []string{"sync", "--source", source, "--dest", dest}
	
	// Apply options
	for _, opt := range options {
		args = opt.apply(args)
	}
	
	return d.executeCommand(args...)
}

func (d *syncDriver) GeneratePlan(source, dest, planFile string, options ...PlanOption) *SyncResult {
	args := []string{"sync", "--source", source, "--dest", dest, "--plan", planFile}
	
	// Apply options
	for _, opt := range options {
		args = opt.apply(args)
	}
	
	return d.executeCommand(args...)
}

func (d *syncDriver) ApplyPlan(planFile string, options ...ApplyOption) *SyncResult {
	args := []string{"sync", "--apply-plan", planFile}
	
	// Apply options
	for _, opt := range options {
		args = opt.apply(args)
	}
	
	return d.executeCommand(args...)
}

func (d *syncDriver) GeneratePatch(source, dest, patchFile string, options ...PatchOption) *SyncResult {
	args := []string{"sync", "--source", source, "--dest", dest, "--patch", patchFile}
	
	// Apply options  
	for _, opt := range options {
		args = opt.apply(args)
	}
	
	return d.executeCommand(args...)
}

func (d *syncDriver) Help() *SyncResult {
	return d.executeCommand("--help")
}

func (d *syncDriver) Version() *SyncResult {
	return d.executeCommand("version")
}

// Option patterns for type-safe configuration

// SyncOption configures sync operations
type SyncOption interface {
	apply(args []string) []string
}

type syncOption struct {
	fn func([]string) []string
}

func (o syncOption) apply(args []string) []string {
	return o.fn(args)
}

// Common sync options
func WithDryRun() SyncOption {
	return syncOption{func(args []string) []string {
		return append(args, "--dry-run")
	}}
}

func WithMode(mode string) SyncOption {
	return syncOption{func(args []string) []string {
		return append(args, "--mode", mode)
	}}
}

func WithInteractive() SyncOption {
	return syncOption{func(args []string) []string {
		return append(args, "--interactive")
	}}
}

// PlanOption configures plan generation
type PlanOption interface {
	apply(args []string) []string
}

type planOption struct {
	fn func([]string) []string
}

func (o planOption) apply(args []string) []string {
	return o.fn(args)
}

func WithIncludeChanges(changes ...string) PlanOption {
	return planOption{func(args []string) []string {
		return append(args, "--include-changes", strings.Join(changes, ","))
	}}
}

func WithExcludeChanges(changes ...string) PlanOption {
	return planOption{func(args []string) []string {
		return append(args, "--exclude-changes", strings.Join(changes, ","))
	}}
}

func WithEditor(editor string) PlanOption {
	return planOption{func(args []string) []string {
		return append(args, "--editor", editor)
	}}
}

// ApplyOption configures plan application
type ApplyOption interface {
	apply(args []string) []string
}

type applyOption struct {
	fn func([]string) []string
}

func (o applyOption) apply(args []string) []string {
	return o.fn(args)
}

func WithConflictStrategy(strategy string) ApplyOption {
	return applyOption{func(args []string) []string {
		return append(args, "--conflict-strategy", strategy)
	}}
}

func WithSkipConflicts() ApplyOption {
	return applyOption{func(args []string) []string {
		return append(args, "--skip-conflicts")
	}}
}

func WithGenerateConflictPlan(planFile string) ApplyOption {
	return applyOption{func(args []string) []string {
		return append(args, "--generate-conflict-plan", planFile)
	}}
}

// PatchOption configures patch generation
type PatchOption interface {
	apply(args []string) []string
}

type patchOption struct {
	fn func([]string) []string
}

func (o patchOption) apply(args []string) []string {
	return o.fn(args)
}

func WithApplyPatch() PatchOption {
	return patchOption{func(args []string) []string {
		return append(args, "--apply-patch")
	}}
}

// Generic command execution methods
func (d *syncDriver) ExecuteCommand(args string) error {
	// Parse the arguments string
	argSlice := strings.Fields(args)
	d.lastResult = d.executeCommand(argSlice...)
	return nil
}

func (d *syncDriver) LastResult() *SyncResult {
	return d.lastResult
}