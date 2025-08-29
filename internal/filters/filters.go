package filters

import (
	"fmt"
	"os"
	"strings"
)

// BuildExcludeFilter creates a temporary filter file for exclude patterns
func BuildExcludeFilter(patterns []string) (string, error) {
	if len(patterns) == 0 {
		return "", nil
	}

	lines := toFilterLines(patterns)
	return writeFilterFile(lines)
}

// BuildOnlyFilter creates a temporary filter file for whitelist (only) mode
func BuildOnlyFilter(onlyPatterns []string) (string, error) {
	if len(onlyPatterns) == 0 {
		return "", nil
	}

	var lines []string

	// For each "only" pattern, we need to:
	// 1. Include parent directories so rsync can traverse to the target
	// 2. Include the pattern itself and its recursive contents
	// 3. Exclude everything else
	for _, pattern := range onlyPatterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}

		// Clean up pattern (remove leading ./ and trailing /**)
		pattern = strings.TrimPrefix(pattern, "./")
		pattern = strings.TrimSuffix(pattern, "/**")
		pattern = strings.TrimSuffix(pattern, "/")

		// Ensure pattern starts with /
		if !strings.HasPrefix(pattern, "/") {
			pattern = "/" + pattern
		}

		// Include root directory
		lines = append(lines, "+ /")

		// Include parent directories
		parts := strings.Split(strings.Trim(pattern, "/"), "/")
		path := ""
		for _, part := range parts {
			path += "/" + part
			lines = append(lines, fmt.Sprintf("+ %s", path))
		}

		// Include the pattern and all its contents recursively
		lines = append(lines, fmt.Sprintf("+ %s/**", pattern))
	}

	// Exclude everything else
	lines = append(lines, "- *")

	return writeFilterFile(lines)
}

// toFilterLines converts patterns to rsync filter lines
func toFilterLines(patterns []string) []string {
	var includes []string
	var excludes []string

	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}

		if strings.HasPrefix(pattern, "!") {
			// Unignore pattern - convert to include rules
			base := strings.TrimPrefix(pattern, "!")
			base = strings.TrimSuffix(base, "/**")
			base = strings.TrimSuffix(base, "/")
			base = ensureSlashPrefix(base)

			// Include root directory
			includes = append(includes, "+ /")

			// Include parent directories
			parts := strings.Split(strings.Trim(base, "/"), "/")
			path := ""
			for _, part := range parts {
				path += "/" + part
				includes = append(includes, fmt.Sprintf("+ %s", path))
			}

			// Include the pattern and all its contents recursively
			includes = append(includes, fmt.Sprintf("+ %s/**", base))
		} else {
			// Regular exclude pattern
			excludes = append(excludes, fmt.Sprintf("- %s", pattern))
		}
	}

	// Combine includes first, then excludes (order matters for rsync)
	var lines []string
	lines = append(lines, includes...)
	lines = append(lines, excludes...)

	return lines
}

// ensureSlashPrefix ensures the path starts with /
func ensureSlashPrefix(path string) string {
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}

// writeFilterFile writes filter lines to a temporary file and returns the filename
func writeFilterFile(lines []string) (string, error) {
	if len(lines) == 0 {
		return "", nil
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "sync-tools-filter-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp filter file: %w", err)
	}
	defer tmpFile.Close()

	// Write filter lines
	for _, line := range lines {
		if _, err := fmt.Fprintln(tmpFile, line); err != nil {
			os.Remove(tmpFile.Name()) // Cleanup on error
			return "", fmt.Errorf("failed to write to temp filter file: %w", err)
		}
	}

	return tmpFile.Name(), nil
}