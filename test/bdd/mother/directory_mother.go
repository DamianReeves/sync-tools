package mother

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TestFile represents a file to be created in a test directory
type TestFile struct {
	Path     string
	Content  string
	Modified time.Time
	Mode     os.FileMode
}

// DirectoryBuilder provides a fluent API for creating test directories
type DirectoryBuilder interface {
	WithFile(path, content string) DirectoryBuilder
	WithFileAt(path, content string, modified time.Time) DirectoryBuilder
	WithDirectory(path string) DirectoryBuilder
	WithMode(mode os.FileMode) DirectoryBuilder
	Build(rootDir string) error
	Files() []TestFile // For inspection
}

type directoryBuilder struct {
	files []TestFile
	defaultMode os.FileMode
}

// NewDirectory creates a new directory builder
func NewDirectory() DirectoryBuilder {
	return &directoryBuilder{
		files: make([]TestFile, 0),
		defaultMode: 0644,
	}
}

func (b *directoryBuilder) WithFile(path, content string) DirectoryBuilder {
	b.files = append(b.files, TestFile{
		Path:     path,
		Content:  content,
		Modified: time.Now(),
		Mode:     b.defaultMode,
	})
	return b
}

func (b *directoryBuilder) WithFileAt(path, content string, modified time.Time) DirectoryBuilder {
	b.files = append(b.files, TestFile{
		Path:     path,
		Content:  content,
		Modified: modified,
		Mode:     b.defaultMode,
	})
	return b
}

func (b *directoryBuilder) WithDirectory(path string) DirectoryBuilder {
	// Directories are represented as files with empty content and different mode
	b.files = append(b.files, TestFile{
		Path:     path,
		Content:  "",
		Modified: time.Now(),
		Mode:     0755 | os.ModeDir,
	})
	return b
}

func (b *directoryBuilder) WithMode(mode os.FileMode) DirectoryBuilder {
	b.defaultMode = mode
	return b
}

func (b *directoryBuilder) Files() []TestFile {
	return b.files
}

func (b *directoryBuilder) Build(rootDir string) error {
	// Ensure root directory exists
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return fmt.Errorf("failed to create root directory %s: %w", rootDir, err)
	}
	
	for _, file := range b.files {
		fullPath := filepath.Join(rootDir, file.Path)
		
		if file.Mode&os.ModeDir != 0 {
			// Create directory
			if err := os.MkdirAll(fullPath, file.Mode&^os.ModeDir); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
			}
		} else {
			// Create parent directories if needed
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory for %s: %w", fullPath, err)
			}
			
			// Create file
			if err := os.WriteFile(fullPath, []byte(file.Content), file.Mode); err != nil {
				return fmt.Errorf("failed to write file %s: %w", fullPath, err)
			}
			
			// Set modification time
			if err := os.Chtimes(fullPath, file.Modified, file.Modified); err != nil {
				return fmt.Errorf("failed to set timestamp for %s: %w", fullPath, err)
			}
		}
	}
	
	return nil
}

// Common test directory scenarios

// BasicSyncScenario creates a basic sync scenario with source-only, dest-only, and conflicting files
func BasicSyncScenario() (DirectoryBuilder, DirectoryBuilder) {
	source := NewDirectory().
		WithFileAt("config/app.yml", "app_config_v1", mustParseTime("2025-08-30T10:30:00")).
		WithFileAt("src/main.js", "console.log('v1')", mustParseTime("2025-08-30T10:35:00")).
		WithFileAt("docs/README.md", "# Version 1", mustParseTime("2025-08-30T09:00:00"))
		
	dest := NewDirectory().
		WithFileAt("config/app.yml", "app_config_v0", mustParseTime("2025-08-30T09:30:00")).
		WithFileAt("config/db.yml", "database_config", mustParseTime("2025-08-30T10:00:00")).
		WithFileAt("docs/README.md", "# Version 2", mustParseTime("2025-08-30T10:00:00"))
	
	return source, dest
}

// ConflictScenario creates a scenario with multiple conflict types
func ConflictScenario() (DirectoryBuilder, DirectoryBuilder) {
	now := time.Now()
	fiveMinAgo := now.Add(-5 * time.Minute)
	tenMinAgo := now.Add(-10 * time.Minute)
	
	source := NewDirectory().
		WithFileAt("newer.txt", "source newer", now).
		WithFileAt("older.txt", "source older", tenMinAgo).
		WithFileAt("same-time.txt", "source content", fiveMinAgo).
		WithFile("source-only.txt", "only in source")
		
	dest := NewDirectory().
		WithFileAt("newer.txt", "dest newer", fiveMinAgo).
		WithFileAt("older.txt", "dest older", now).
		WithFileAt("same-time.txt", "dest content", fiveMinAgo).
		WithFile("dest-only.txt", "only in dest")
	
	return source, dest
}

// EmptyDirectories creates empty source and dest directories
func EmptyDirectories() (DirectoryBuilder, DirectoryBuilder) {
	return NewDirectory(), NewDirectory()
}

// SourceOnlyScenario creates a scenario with files only in source
func SourceOnlyScenario() (DirectoryBuilder, DirectoryBuilder) {
	source := NewDirectory().
		WithFile("file1.txt", "content 1").
		WithFile("sub/file2.txt", "content 2").
		WithFile("another/deep/file3.txt", "content 3")
		
	dest := NewDirectory()
	
	return source, dest
}

// IdenticalDirectories creates identical source and dest directories
func IdenticalDirectories() (DirectoryBuilder, DirectoryBuilder) {
	timestamp := mustParseTime("2025-08-30T10:00:00")
	
	source := NewDirectory().
		WithFileAt("same1.txt", "identical content", timestamp).
		WithFileAt("same2.txt", "also identical", timestamp)
		
	dest := NewDirectory().
		WithFileAt("same1.txt", "identical content", timestamp).
		WithFileAt("same2.txt", "also identical", timestamp)
	
	return source, dest
}

// mustParseTime parses a time string in the test format, panicking on error
func mustParseTime(timeStr string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05", timeStr)
	if err != nil {
		panic(fmt.Sprintf("failed to parse time %s: %v", timeStr, err))
	}
	return t
}

// ParseTestTime parses a time string in the test format, returning error
func ParseTestTime(timeStr string) (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05", timeStr)
}