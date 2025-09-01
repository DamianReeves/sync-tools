package wizard

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DirectoryTree represents a file system tree browser
type DirectoryTree struct {
	root         string
	currentPath  string
	items        []DirectoryItem
	selectedIdx  int
	expanded     map[string]bool
	scrollOffset int
	viewHeight   int
}

// DirectoryItem represents an item in the directory tree
type DirectoryItem struct {
	Name     string
	Path     string
	IsDir    bool
	Level    int
	Size     int64
	FileCount int
	IsExpanded bool
	Parent   string
}

// NewDirectoryTree creates a new directory tree browser
func NewDirectoryTree(root string) *DirectoryTree {
	if root == "" {
		root = "/"
	}
	
	return &DirectoryTree{
		root:        root,
		currentPath: root,
		expanded:    make(map[string]bool),
		viewHeight:  10, // Default height
	}
}

// SetViewHeight sets the visible height of the tree
func (dt *DirectoryTree) SetViewHeight(height int) {
	dt.viewHeight = height
}

// GetCurrentPath returns the currently selected path
func (dt *DirectoryTree) GetCurrentPath() string {
	if dt.selectedIdx < len(dt.items) {
		return dt.items[dt.selectedIdx].Path
	}
	return dt.currentPath
}

// SetCurrentPath sets the current path and rebuilds the tree
func (dt *DirectoryTree) SetCurrentPath(path string) error {
	if _, err := os.Stat(path); err != nil {
		return err
	}
	
	dt.currentPath = path
	dt.selectedIdx = 0
	dt.scrollOffset = 0
	return dt.Refresh()
}

// Refresh rebuilds the directory tree
func (dt *DirectoryTree) Refresh() error {
	dt.items = nil
	return dt.buildTree(dt.root, 0)
}

// buildTree recursively builds the directory tree
func (dt *DirectoryTree) buildTree(path string, level int) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	// Sort entries: directories first, then by name
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
	})

	for _, entry := range entries {
		// Skip hidden files unless explicitly shown
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		fullPath := filepath.Join(path, entry.Name())
		item := DirectoryItem{
			Name:   entry.Name(),
			Path:   fullPath,
			IsDir:  entry.IsDir(),
			Level:  level,
			Parent: path,
		}

		// Get size and file count for directories
		if entry.IsDir() {
			item.FileCount, item.Size = dt.getDirStats(fullPath)
			item.IsExpanded = dt.expanded[fullPath]
		} else {
			if info, err := entry.Info(); err == nil {
				item.Size = info.Size()
			}
		}

		dt.items = append(dt.items, item)

		// Recursively add subdirectories if expanded
		if entry.IsDir() && dt.expanded[fullPath] {
			dt.buildTree(fullPath, level+1)
		}
	}

	return nil
}

// getDirStats returns file count and total size for a directory
func (dt *DirectoryTree) getDirStats(dirPath string) (int, int64) {
	var fileCount int
	var totalSize int64

	filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip inaccessible files
		}
		if !d.IsDir() {
			fileCount++
			if info, err := d.Info(); err == nil {
				totalSize += info.Size()
			}
		}
		return nil
	})

	return fileCount, totalSize
}

// MoveUp moves the selection up
func (dt *DirectoryTree) MoveUp() {
	if dt.selectedIdx > 0 {
		dt.selectedIdx--
		dt.adjustScroll()
	}
}

// MoveDown moves the selection down
func (dt *DirectoryTree) MoveDown() {
	if dt.selectedIdx < len(dt.items)-1 {
		dt.selectedIdx++
		dt.adjustScroll()
	}
}

// ExpandSelected expands the currently selected directory
func (dt *DirectoryTree) ExpandSelected() error {
	if dt.selectedIdx >= len(dt.items) {
		return nil
	}

	item := dt.items[dt.selectedIdx]
	if !item.IsDir {
		return nil
	}

	dt.expanded[item.Path] = true
	return dt.Refresh()
}

// CollapseSelected collapses the currently selected directory
func (dt *DirectoryTree) CollapseSelected() error {
	if dt.selectedIdx >= len(dt.items) {
		return nil
	}

	item := dt.items[dt.selectedIdx]
	if !item.IsDir {
		return nil
	}

	dt.expanded[item.Path] = false
	return dt.Refresh()
}

// SelectPath selects a specific path in the tree
func (dt *DirectoryTree) SelectPath(path string) error {
	// First, ensure all parent directories are expanded
	parent := filepath.Dir(path)
	for parent != dt.root && parent != "/" {
		dt.expanded[parent] = true
		parent = filepath.Dir(parent)
	}

	// Refresh to show expanded items
	if err := dt.Refresh(); err != nil {
		return err
	}

	// Find and select the path
	for i, item := range dt.items {
		if item.Path == path {
			dt.selectedIdx = i
			dt.adjustScroll()
			return nil
		}
	}

	return fmt.Errorf("path not found: %s", path)
}

// adjustScroll adjusts scroll offset to keep selection visible
func (dt *DirectoryTree) adjustScroll() {
	if dt.selectedIdx < dt.scrollOffset {
		dt.scrollOffset = dt.selectedIdx
	} else if dt.selectedIdx >= dt.scrollOffset+dt.viewHeight {
		dt.scrollOffset = dt.selectedIdx - dt.viewHeight + 1
	}
}

// Render renders the directory tree as a string
func (dt *DirectoryTree) Render() string {
	if len(dt.items) == 0 {
		return "No items to display"
	}

	var lines []string
	end := dt.scrollOffset + dt.viewHeight
	if end > len(dt.items) {
		end = len(dt.items)
	}

	for i := dt.scrollOffset; i < end; i++ {
		item := dt.items[i]
		line := dt.renderItem(item, i == dt.selectedIdx)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// renderItem renders a single directory item
func (dt *DirectoryTree) renderItem(item DirectoryItem, isSelected bool) string {
	// Create indentation
	indent := strings.Repeat("  ", item.Level)
	
	// Icon and name
	icon := "📁"
	if !item.IsDir {
		icon = "📄"
	}
	
	name := item.Name
	if isSelected {
		name = fmt.Sprintf("● %s", name)
	} else {
		name = fmt.Sprintf("  %s", name)
	}

	// Size information for directories
	sizeInfo := ""
	if item.IsDir && item.FileCount > 0 {
		sizeInfo = fmt.Sprintf(" (%d files, %s)", item.FileCount, formatSize(item.Size))
	} else if !item.IsDir {
		sizeInfo = fmt.Sprintf(" (%s)", formatSize(item.Size))
	}

	return fmt.Sprintf("%s%s %s%s", indent, icon, name, sizeInfo)
}

// formatSize formats bytes into human-readable format
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	if bytes >= TB {
		return fmt.Sprintf("%.1f TB", float64(bytes)/TB)
	} else if bytes >= GB {
		return fmt.Sprintf("%.1f GB", float64(bytes)/GB)
	} else if bytes >= MB {
		return fmt.Sprintf("%.1f MB", float64(bytes)/MB)
	} else if bytes >= KB {
		return fmt.Sprintf("%.1f KB", float64(bytes)/KB)
	}
	return fmt.Sprintf("%d B", bytes)
}

// GetVisibleItems returns the currently visible items
func (dt *DirectoryTree) GetVisibleItems() []DirectoryItem {
	end := dt.scrollOffset + dt.viewHeight
	if end > len(dt.items) {
		end = len(dt.items)
	}
	return dt.items[dt.scrollOffset:end]
}

// GetSelectedItem returns the currently selected item
func (dt *DirectoryTree) GetSelectedItem() *DirectoryItem {
	if dt.selectedIdx < len(dt.items) {
		return &dt.items[dt.selectedIdx]
	}
	return nil
}