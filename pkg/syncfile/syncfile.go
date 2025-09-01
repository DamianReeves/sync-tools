package syncfile

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/DamianReeves/sync-tools/internal/rsync"
)

// SyncFile represents a parsed SyncFile (Dockerfile-like format for sync operations)
type SyncFile struct {
	Instructions []Instruction
	Variables    map[string]string
}

// InstructionType represents the type of SyncFile instruction
type InstructionType string

const (
	// Core sync instructions
	InstSync    InstructionType = "SYNC"    // SYNC source dest [OPTIONS]
	InstExclude InstructionType = "EXCLUDE" // EXCLUDE pattern
	InstInclude InstructionType = "INCLUDE" // INCLUDE pattern (unignore)
	InstOnly    InstructionType = "ONLY"    // ONLY pattern (whitelist mode)

	// Configuration instructions
	InstMode         InstructionType = "MODE"       // MODE one-way|two-way
	InstDryRun       InstructionType = "DRYRUN"     // DRYRUN true|false
	InstUseGitignore InstructionType = "GITIGNORE"  // GITIGNORE true|false
	InstHiddenDirs   InstructionType = "HIDDENDIRS" // HIDDENDIRS exclude|include

	// Patch instructions
	InstPatch       InstructionType = "PATCH"       // PATCH filename
	InstApplyPatch  InstructionType = "APPLYPATCH"  // APPLYPATCH true|false
	InstPreview     InstructionType = "PREVIEW"     // PREVIEW true|false
	InstAutoConfirm InstructionType = "AUTOCONFIRM" // AUTOCONFIRM true|false (like -y flag)

	// Post-sync action instructions
	InstAppend  InstructionType = "APPEND"  // APPEND filename: content END APPEND
	InstPrepend InstructionType = "PREPEND" // PREPEND filename: content END PREPEND

	// Variable and environment instructions
	InstVar InstructionType = "VAR" // VAR name=value
	InstEnv InstructionType = "ENV" // ENV name=value (exported to rsync)

	// Advanced instructions
	InstRun     InstructionType = "RUN"     // RUN command (pre/post sync hooks)
	InstComment InstructionType = "COMMENT" // # Comment
)

// Instruction represents a single SyncFile instruction
type Instruction struct {
	Type          InstructionType
	Args          []string
	Comment       string
	LineNum       int
	InlineContent string // For multi-line instructions like APPEND
}

// ParseSyncFile parses a SyncFile from the given path
func ParseSyncFile(path string) (*SyncFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open SyncFile: %w", err)
	}
	defer file.Close()

	sf := &SyncFile{
		Instructions: make([]Instruction, 0),
		Variables:    make(map[string]string),
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Handle comments
		if strings.HasPrefix(line, "#") {
			sf.Instructions = append(sf.Instructions, Instruction{
				Type:    InstComment,
				Comment: line[1:],
				LineNum: lineNum,
			})
			continue
		}

		// Check for APPEND inline block
		if strings.HasPrefix(strings.ToUpper(line), "APPEND ") && strings.HasSuffix(line, ":") {
			instruction, err := parseAppendBlock(scanner, line, &lineNum)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNum, err)
			}
			sf.Instructions = append(sf.Instructions, instruction)
			continue
		}

		// Check for PREPEND inline block
		if strings.HasPrefix(strings.ToUpper(line), "PREPEND ") && strings.HasSuffix(line, ":") {
			instruction, err := parsePrependBlock(scanner, line, &lineNum)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNum, err)
			}
			sf.Instructions = append(sf.Instructions, instruction)
			continue
		}

		// Parse regular instruction
		instruction, err := parseInstruction(line, lineNum)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}

		sf.Instructions = append(sf.Instructions, instruction)

		// Handle variable assignments
		if instruction.Type == InstVar || instruction.Type == InstEnv {
			if len(instruction.Args) > 0 {
				parts := strings.SplitN(instruction.Args[0], "=", 2)
				if len(parts) == 2 {
					sf.Variables[parts[0]] = expandVariables(parts[1], sf.Variables)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading SyncFile: %w", err)
	}

	return sf, nil
}

// parseInstruction parses a single instruction line
func parseInstruction(line string, lineNum int) (Instruction, error) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return Instruction{}, fmt.Errorf("empty instruction")
	}

	instType := InstructionType(strings.ToUpper(parts[0]))
	args := parts[1:]

	// Validate instruction type and arguments
	switch instType {
	case InstSync:
		if len(args) < 2 {
			return Instruction{}, fmt.Errorf("SYNC requires at least 2 arguments: source dest")
		}
	case InstExclude, InstInclude, InstOnly:
		if len(args) < 1 {
			return Instruction{}, fmt.Errorf("%s requires at least 1 argument", instType)
		}
	case InstMode:
		if len(args) != 1 || (args[0] != "one-way" && args[0] != "two-way") {
			return Instruction{}, fmt.Errorf("MODE must be 'one-way' or 'two-way'")
		}
	case InstDryRun, InstUseGitignore, InstApplyPatch, InstPreview, InstAutoConfirm:
		if len(args) != 1 || (args[0] != "true" && args[0] != "false") {
			return Instruction{}, fmt.Errorf("%s must be 'true' or 'false'", instType)
		}
	case InstPatch:
		if len(args) != 1 {
			return Instruction{}, fmt.Errorf("PATCH requires exactly 1 argument: filename")
		}
	case InstHiddenDirs:
		if len(args) != 1 || (args[0] != "exclude" && args[0] != "include") {
			return Instruction{}, fmt.Errorf("HIDDENDIRS must be 'exclude' or 'include'")
		}
	case InstVar, InstEnv:
		if len(args) != 1 || !strings.Contains(args[0], "=") {
			return Instruction{}, fmt.Errorf("%s requires format: name=value", instType)
		}
	case InstRun:
		if len(args) < 1 {
			return Instruction{}, fmt.Errorf("RUN requires at least 1 argument")
		}
	case InstAppend:
		if len(args) < 1 {
			return Instruction{}, fmt.Errorf("APPEND requires at least 1 argument")
		}
		// Validate --file flag format: APPEND --file source.txt target.txt
		if len(args) >= 1 && args[0] == "--file" && len(args) != 3 {
			return Instruction{}, fmt.Errorf("APPEND --file requires format: APPEND --file source.txt target.txt")
		}
	case InstPrepend:
		if len(args) < 1 {
			return Instruction{}, fmt.Errorf("PREPEND requires at least 1 argument")
		}
		// Validate --file flag format: PREPEND --file source.txt target.txt
		if len(args) >= 1 && args[0] == "--file" && len(args) != 3 {
			return Instruction{}, fmt.Errorf("PREPEND --file requires format: PREPEND --file source.txt target.txt")
		}
	default:
		return Instruction{}, fmt.Errorf("unknown instruction: %s", instType)
	}

	return Instruction{
		Type:    instType,
		Args:    args,
		LineNum: lineNum,
	}, nil
}

// parseAppendBlock parses an APPEND inline block with content until END APPEND
func parseAppendBlock(scanner *bufio.Scanner, firstLine string, lineNum *int) (Instruction, error) {
	startLineNum := *lineNum

	// Parse the APPEND line: "APPEND [flags] filename:"
	line := strings.TrimSuffix(firstLine, ":")
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return Instruction{}, fmt.Errorf("APPEND requires format: APPEND [flags] filename:")
	}

	// Find the filename (last part) and collect flags
	filename := parts[len(parts)-1]
	flags := parts[1 : len(parts)-1] // Everything between APPEND and filename

	// Combine flags and filename for args
	args := append(flags, filename)
	var contentLines []string

	// Read lines until END APPEND
	for scanner.Scan() {
		*lineNum++
		line := strings.TrimSpace(scanner.Text())

		if strings.ToUpper(line) == "END APPEND" {
			// Found the end marker
			break
		}

		// Add the raw line (preserving original spacing for content)
		contentLines = append(contentLines, scanner.Text())
	}

	if *lineNum == startLineNum {
		return Instruction{}, fmt.Errorf("APPEND block not closed with END APPEND")
	}

	// Join content with newlines and normalize indentation
	content := strings.Join(contentLines, "\n")
	content = normalizeIndentation(content)

	return Instruction{
		Type:          InstAppend,
		Args:          args,
		LineNum:       startLineNum,
		InlineContent: content,
	}, nil
}

// parsePrependBlock parses a PREPEND inline block with content until END PREPEND
func parsePrependBlock(scanner *bufio.Scanner, firstLine string, lineNum *int) (Instruction, error) {
	startLineNum := *lineNum

	// Parse the PREPEND line: "PREPEND [flags] filename:"
	line := strings.TrimSuffix(firstLine, ":")
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return Instruction{}, fmt.Errorf("PREPEND requires format: PREPEND [flags] filename:")
	}

	// Find the filename (last part) and collect flags
	filename := parts[len(parts)-1]
	flags := parts[1 : len(parts)-1] // Everything between PREPEND and filename

	// Combine flags and filename for args
	args := append(flags, filename)
	var contentLines []string

	// Read lines until END PREPEND
	for scanner.Scan() {
		*lineNum++
		line := strings.TrimSpace(scanner.Text())

		if strings.ToUpper(line) == "END PREPEND" {
			// Found the end marker
			break
		}

		// Add the raw line (preserving original spacing for content)
		contentLines = append(contentLines, scanner.Text())
	}

	if *lineNum == startLineNum {
		return Instruction{}, fmt.Errorf("PREPEND block not closed with END PREPEND")
	}

	// Join content with newlines and normalize indentation
	content := strings.Join(contentLines, "\n")
	content = normalizeIndentation(content)

	return Instruction{
		Type:          InstPrepend,
		Args:          args,
		LineNum:       startLineNum,
		InlineContent: content,
	}, nil
}

// normalizeIndentation removes common leading whitespace from multi-line content
func normalizeIndentation(content string) string {
	if content == "" {
		return content
	}

	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return content
	}

	// Find minimum indentation (ignoring empty lines)
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue // Skip empty lines
		}

		indent := 0
		for _, char := range line {
			if char == ' ' || char == '\t' {
				if char == '\t' {
					indent += 4 // Count tabs as 4 spaces
				} else {
					indent++
				}
			} else {
				break
			}
		}

		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	// If no indentation found, return as-is
	if minIndent <= 0 {
		return content
	}

	// Remove common indentation from all lines
	var normalizedLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			// Preserve empty lines as-is
			normalizedLines = append(normalizedLines, "")
		} else {
			// Remove common indentation
			dedented := removeLeadingSpaces(line, minIndent)
			normalizedLines = append(normalizedLines, dedented)
		}
	}

	return strings.Join(normalizedLines, "\n")
}

// removeLeadingSpaces removes up to n leading spaces/tabs from a line
func removeLeadingSpaces(line string, n int) string {
	removed := 0
	for i, char := range line {
		if removed >= n {
			return line[i:]
		}

		if char == ' ' {
			removed++
		} else if char == '\t' {
			removed += 4 // Count tabs as 4 spaces
			if removed > n {
				// If we removed too much, put back partial spaces
				return strings.Repeat(" ", removed-n) + line[i+1:]
			}
		} else {
			break
		}
	}

	return line[len(line):] // Return empty string if all chars were whitespace
}

// expandVariables expands variable references in a string
func expandVariables(s string, vars map[string]string) string {
	result := s
	for name, value := range vars {
		result = strings.ReplaceAll(result, "${"+name+"}", value)
		result = strings.ReplaceAll(result, "$"+name, value)
	}
	return result
}

// parseAppendAction converts an APPEND instruction to a PostSyncAction
func parseAppendAction(inst Instruction, variables map[string]string) rsync.PostSyncAction {
	action := rsync.PostSyncAction{
		Type:    rsync.PostSyncAppend,
		Flags:   make([]string, 0),
		Content: inst.InlineContent,
	}

	// Parse arguments: [flags...] filename
	args := inst.Args
	if len(args) == 0 {
		return action
	}

	// Last argument is always the target file
	targetFile := expandVariables(args[len(args)-1], variables)
	action.TargetFile = targetFile

	// Everything before the last argument are flags
	flags := args[:len(args)-1]

	// Parse flags
	for _, flag := range flags {
		switch flag {
		case "--file":
			// For --file flag, we need the source file from the next argument
			// This should have been validated during parsing
			if len(args) >= 3 && args[0] == "--file" {
				action.SourceFile = expandVariables(args[1], variables)
				action.TargetFile = expandVariables(args[2], variables)
				action.Content = "" // No inline content for file-based append
			}
		default:
			// Add other flags as-is
			action.Flags = append(action.Flags, flag)
		}
	}

	return action
}

// parsePrependAction converts a PREPEND instruction to a PostSyncAction
func parsePrependAction(inst Instruction, variables map[string]string) rsync.PostSyncAction {
	action := rsync.PostSyncAction{
		Type:    rsync.PostSyncPrepend,
		Flags:   make([]string, 0),
		Content: inst.InlineContent,
	}

	// Parse arguments: [flags...] filename
	args := inst.Args
	if len(args) == 0 {
		return action
	}

	// Last argument is always the target file
	targetFile := expandVariables(args[len(args)-1], variables)
	action.TargetFile = targetFile

	// Everything before the last argument are flags
	flags := args[:len(args)-1]

	// Parse flags
	for _, flag := range flags {
		switch flag {
		case "--file":
			// For --file flag, we need the source file from the next argument
			// This should have been validated during parsing
			if len(args) >= 3 && args[0] == "--file" {
				action.SourceFile = expandVariables(args[1], variables)
				action.TargetFile = expandVariables(args[2], variables)
				action.Content = "" // No inline content for file-based prepend
			}
		default:
			// Add other flags as-is
			action.Flags = append(action.Flags, flag)
		}
	}

	return action
}

// ToRsyncOptions converts a SyncFile to rsync.Options
func (sf *SyncFile) ToRsyncOptions() ([]*rsync.Options, error) {
	var optsList []*rsync.Options
	var currentOpts *rsync.Options

	for _, inst := range sf.Instructions {
		switch inst.Type {
		case InstSync:
			// Start a new sync operation
			if currentOpts != nil {
				optsList = append(optsList, currentOpts)
			}

			source := expandVariables(inst.Args[0], sf.Variables)
			dest := expandVariables(inst.Args[1], sf.Variables)

			// Resolve paths
			if !filepath.IsAbs(source) {
				source = filepath.Join(".", source)
			}
			if !filepath.IsAbs(dest) {
				dest = filepath.Join(".", dest)
			}

			currentOpts = &rsync.Options{
				Source: source,
				Dest:   dest,
				Mode:   "one-way", // default
			}

			// Process additional options in SYNC command
			for i := 2; i < len(inst.Args); i++ {
				option := inst.Args[i]
				switch option {
				case "--dry-run":
					currentOpts.DryRun = true
				case "--two-way":
					currentOpts.Mode = "two-way"
				case "--use-gitignore":
					currentOpts.UseSourceGitignore = true
				case "--exclude-hidden":
					currentOpts.ExcludeHiddenDirs = true
				}
			}

		case InstMode:
			if currentOpts != nil {
				currentOpts.Mode = inst.Args[0]
			}

		case InstDryRun:
			if currentOpts != nil {
				dryRun, _ := strconv.ParseBool(inst.Args[0])
				currentOpts.DryRun = dryRun
			}

		case InstUseGitignore:
			if currentOpts != nil {
				useGitignore, _ := strconv.ParseBool(inst.Args[0])
				currentOpts.UseSourceGitignore = useGitignore
			}

		case InstHiddenDirs:
			if currentOpts != nil {
				currentOpts.ExcludeHiddenDirs = (inst.Args[0] == "exclude")
			}

		case InstExclude:
			if currentOpts != nil {
				pattern := expandVariables(inst.Args[0], sf.Variables)
				currentOpts.IgnoreSrc = append(currentOpts.IgnoreSrc, pattern)
			}

		case InstInclude:
			if currentOpts != nil {
				pattern := expandVariables(inst.Args[0], sf.Variables)
				// Include patterns are prefixed with !
				currentOpts.IgnoreSrc = append(currentOpts.IgnoreSrc, "!"+pattern)
			}

		case InstOnly:
			if currentOpts != nil {
				pattern := expandVariables(inst.Args[0], sf.Variables)
				currentOpts.Only = append(currentOpts.Only, pattern)
			}

		case InstPatch:
			if currentOpts != nil {
				patchFile := expandVariables(inst.Args[0], sf.Variables)
				currentOpts.Patch = patchFile
			}

		case InstApplyPatch:
			if currentOpts != nil {
				applyPatch, _ := strconv.ParseBool(inst.Args[0])
				currentOpts.ApplyPatch = applyPatch
			}

		case InstPreview:
			if currentOpts != nil {
				preview, _ := strconv.ParseBool(inst.Args[0])
				currentOpts.Preview = preview
			}

		case InstAutoConfirm:
			if currentOpts != nil {
				autoConfirm, _ := strconv.ParseBool(inst.Args[0])
				currentOpts.Yes = autoConfirm
			}

		case InstAppend:
			if currentOpts != nil {
				// Parse APPEND instruction and add to post-sync actions
				action := parseAppendAction(inst, sf.Variables)
				currentOpts.PostSyncActions = append(currentOpts.PostSyncActions, action)
			}

		case InstPrepend:
			if currentOpts != nil {
				// Parse PREPEND instruction and add to post-sync actions
				action := parsePrependAction(inst, sf.Variables)
				currentOpts.PostSyncActions = append(currentOpts.PostSyncActions, action)
			}
		}
	}

	// Add the last sync operation
	if currentOpts != nil {
		optsList = append(optsList, currentOpts)
	}

	if len(optsList) == 0 {
		return nil, fmt.Errorf("no SYNC instructions found in SyncFile")
	}

	return optsList, nil
}

// Example SyncFile content:
/*
# SyncFile - Docker-like syntax for sync operations
# This is a multi-project sync configuration

VAR PROJECT_ROOT=/home/user/projects
VAR BACKUP_ROOT=/backup

# Preview changes before syncing
SYNC ${PROJECT_ROOT}/docs ${BACKUP_ROOT}/docs
MODE one-way
PREVIEW true
EXCLUDE *.tmp
EXCLUDE .DS_Store
INCLUDE !important.tmp

# Generate patch file for source code changes
SYNC ${PROJECT_ROOT}/src ${BACKUP_ROOT}/src
MODE two-way
PATCH src-changes.patch
APPLYPATCH true
AUTOCONFIRM false  # Will prompt for confirmation
GITIGNORE true
HIDDENDIRS exclude
ONLY *.go
ONLY *.py
ONLY *.js

# Sync config files with auto-applied patch
SYNC ${PROJECT_ROOT}/config ${BACKUP_ROOT}/config
PATCH config-update.patch
APPLYPATCH true
AUTOCONFIRM true  # Like -y flag, no confirmation prompt
EXCLUDE secrets/
INCLUDE !config/main.conf

# Traditional sync without patches
SYNC ${PROJECT_ROOT}/data ${BACKUP_ROOT}/data
DRYRUN false
EXCLUDE cache/
*/
