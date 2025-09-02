package wizard

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// DirectoryBrowser provides filesystem browsing functionality
type DirectoryBrowser struct {
	currentPath   string
	entries       []DirectoryEntry
	selectedIndex int
}

// DirectoryEntry represents a filesystem entry
type DirectoryEntry struct {
	Name      string
	Path      string
	IsDir     bool
	Size      int64
	FileCount int
}

// NewDirectoryBrowser creates a new directory browser
func NewDirectoryBrowser(startPath string) *DirectoryBrowser {
	browser := &DirectoryBrowser{
		currentPath:   startPath,
		selectedIndex: 0,
	}
	browser.refreshEntries()
	return browser
}

// GetCurrentPath returns the current directory path
func (db *DirectoryBrowser) GetCurrentPath() string {
	return db.currentPath
}

// GetEntries returns the current directory entries
func (db *DirectoryBrowser) GetEntries() []DirectoryEntry {
	return db.entries
}

// GetSelectedIndex returns the currently selected entry index
func (db *DirectoryBrowser) GetSelectedIndex() int {
	return db.selectedIndex
}

// GetSelectedEntry returns the currently selected entry
func (db *DirectoryBrowser) GetSelectedEntry() *DirectoryEntry {
	if db.selectedIndex >= 0 && db.selectedIndex < len(db.entries) {
		return &db.entries[db.selectedIndex]
	}
	return nil
}

// MoveUp moves selection up
func (db *DirectoryBrowser) MoveUp() {
	if db.selectedIndex > 0 {
		db.selectedIndex--
	}
}

// MoveDown moves selection down
func (db *DirectoryBrowser) MoveDown() {
	if db.selectedIndex < len(db.entries)-1 {
		db.selectedIndex++
	}
}

// EnterDirectory enters the selected directory
func (db *DirectoryBrowser) EnterDirectory() bool {
	entry := db.GetSelectedEntry()
	if entry == nil || !entry.IsDir {
		return false
	}

	db.currentPath = entry.Path
	db.selectedIndex = 0
	db.refreshEntries()
	return true
}

// GoUp goes to parent directory
func (db *DirectoryBrowser) GoUp() bool {
	parent := filepath.Dir(db.currentPath)
	if parent == db.currentPath {
		return false // Already at root
	}

	db.currentPath = parent
	db.selectedIndex = 0
	db.refreshEntries()
	return true
}

// SetPath sets the current path and refreshes entries
func (db *DirectoryBrowser) SetPath(path string) error {
	// Validate path exists
	if _, err := os.Stat(path); err != nil {
		return err
	}

	db.currentPath = path
	db.selectedIndex = 0
	db.refreshEntries()
	return nil
}

// refreshEntries scans the current directory and populates entries
func (db *DirectoryBrowser) refreshEntries() {
	db.entries = []DirectoryEntry{}

	// Add parent directory entry if not at root
	if parent := filepath.Dir(db.currentPath); parent != db.currentPath {
		db.entries = append(db.entries, DirectoryEntry{
			Name:  "..",
			Path:  parent,
			IsDir: true,
			Size:  0,
		})
	}

	// Read directory contents
	files, err := os.ReadDir(db.currentPath)
	if err != nil {
		return
	}

	// Process entries
	var dirEntries []DirectoryEntry
	var fileEntries []DirectoryEntry

	for _, file := range files {
		fullPath := filepath.Join(db.currentPath, file.Name())

		entry := DirectoryEntry{
			Name:  file.Name(),
			Path:  fullPath,
			IsDir: file.IsDir(),
		}

		// Get size and file count for directories
		if file.IsDir() {
			entry.FileCount, entry.Size = db.getDirectoryStats(fullPath)
			dirEntries = append(dirEntries, entry)
		} else {
			if info, err := file.Info(); err == nil {
				entry.Size = info.Size()
			}
			fileEntries = append(fileEntries, entry)
		}
	}

	// Sort directories and files separately
	sort.Slice(dirEntries, func(i, j int) bool {
		return dirEntries[i].Name < dirEntries[j].Name
	})
	sort.Slice(fileEntries, func(i, j int) bool {
		return fileEntries[i].Name < fileEntries[j].Name
	})

	// Add directories first, then files
	db.entries = append(db.entries, dirEntries...)
	db.entries = append(db.entries, fileEntries...)
}

// getDirectoryStats calculates file count and total size for a directory
func (db *DirectoryBrowser) getDirectoryStats(dirPath string) (fileCount int, totalSize int64) {
	_ = filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if !d.IsDir() {
			fileCount++
			if info, err := d.Info(); err == nil {
				totalSize += info.Size()
			}
		}
		return nil
	})
	return
}

// FormatSize formats a byte size as human readable string
func FormatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return "0B"
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(size)/float64(div), "KMGTPE"[exp])
}
