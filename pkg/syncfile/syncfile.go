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
	InstSync        InstructionType = "SYNC"        // SYNC source dest [OPTIONS]
	InstExclude     InstructionType = "EXCLUDE"     // EXCLUDE pattern
	InstInclude     InstructionType = "INCLUDE"     // INCLUDE pattern (unignore)
	InstOnly        InstructionType = "ONLY"        // ONLY pattern (whitelist mode)
	
	// Configuration instructions
	InstMode        InstructionType = "MODE"        // MODE one-way|two-way
	InstDryRun      InstructionType = "DRYRUN"      // DRYRUN true|false
	InstUseGitignore InstructionType = "GITIGNORE"  // GITIGNORE true|false
	InstHiddenDirs  InstructionType = "HIDDENDIRS"  // HIDDENDIRS exclude|include
	
	// Variable and environment instructions
	InstVar         InstructionType = "VAR"         // VAR name=value
	InstEnv         InstructionType = "ENV"         // ENV name=value (exported to rsync)
	
	// Advanced instructions
	InstRun         InstructionType = "RUN"         // RUN command (pre/post sync hooks)
	InstComment     InstructionType = "COMMENT"     // # Comment
)

// Instruction represents a single SyncFile instruction
type Instruction struct {
	Type     InstructionType
	Args     []string
	Comment  string
	LineNum  int
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

		// Parse instruction
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
	case InstDryRun, InstUseGitignore:
		if len(args) != 1 || (args[0] != "true" && args[0] != "false") {
			return Instruction{}, fmt.Errorf("%s must be 'true' or 'false'", instType)
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
	default:
		return Instruction{}, fmt.Errorf("unknown instruction: %s", instType)
	}

	return Instruction{
		Type:    instType,
		Args:    args,
		LineNum: lineNum,
	}, nil
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

# Sync documentation
SYNC ${PROJECT_ROOT}/docs ${BACKUP_ROOT}/docs --dry-run
MODE one-way
EXCLUDE *.tmp
EXCLUDE .DS_Store
INCLUDE !important.tmp

# Sync source code
SYNC ${PROJECT_ROOT}/src ${BACKUP_ROOT}/src
MODE two-way
GITIGNORE true
HIDDENDIRS exclude
ONLY *.go
ONLY *.py
ONLY *.js

# Sync config files
SYNC ${PROJECT_ROOT}/config ${BACKUP_ROOT}/config
DRYRUN false
EXCLUDE secrets/
INCLUDE !config/main.conf
*/